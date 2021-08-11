package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

const (
	FullnodeAPIFile     = "/etc/profile.d/fullnode-api-info.sh"
	FullnodeEnvKey      = "FULLNODE_API_INFO"
	FullnodeServiceFile = "/etc/systemd/system/lotus-daemon.service"
	MinerAPIFile        = "/etc/profile.d/miner-api-info.sh"
	MinerEnvKey         = "MINER_API_INFO"
	MinerServiceFile    = "/etc/systemd/system/lotus-miner.service"
	WorkerServiceFile   = "/etc/systemd/system/lotus-worker.service"
	CommonSetupFile     = "/etc/profile.d/lotus-setup.sh"
	StorageListDescFile = "storage.json"
	StorageMetaFile     = "sectorstore.json"
	ProcSelfMounts      = "/proc/self/mounts"
	HostsFile           = "/etc/hosts"
	CephConfigFile      = "/etc/ceph/ceph.conf"
)

type nodeDesc struct {
	apiInfo string
	ip      string
}

type Parser struct {
	fileAPIInfo            map[string]nodeDesc
	minerStorageChilds     []string
	storagePath            string
	validStoragePath       bool
	storageConfig          StorageConfig
	validStorageConfig     bool
	storageMetas           []LocalStorageMeta
	cephEntries            map[string]struct{}
	localAddr              string
	cephStoragePeers       map[string]string
	storageSubRole         string
	storageChilds          []string
	minerLogFile           string
	fullnodeLogFile        string
	minerShareStorageRoot  string
	chiaMinerNodeLogFile   string
	chiaPlotterLogFile     string
	minerApiHost           string
	fullnodeApiHost        string
	minerRepoDirApiFile    string
	fullnodeRepoDirApiFile string
}

type OSSInfo struct {
	URL        string
	AccessKey  string
	SecretKey  string
	BucketName string
	Prefix     string
	CanWrite   bool
}

// LocalStorageMeta [path]/sectorstore.json
type LocalStorageMeta struct {
	ID       uuid.UUID
	Weight   uint64 // 0 = readonly
	CanSeal  bool
	CanStore bool
	Oss      bool
	OssInfo  OSSInfo
}

// StorageConfig .lotusstorage/storage.json
type StorageConfig struct {
	StoragePaths []LocalPath
}

type LocalPath struct {
	Path string
}

func NewParser() *Parser {
	parser := &Parser{
		fileAPIInfo:           map[string]nodeDesc{},
		cephEntries:           map[string]struct{}{},
		cephStoragePeers:      map[string]string{},
		minerShareStorageRoot: "/opt/sharestorage",
		chiaMinerNodeLogFile:  "/var/log/chia/miner.log",
		chiaPlotterLogFile:    "/var/log/chia-plotter.log",
	}
	err := parser.parse()
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse node: %v", err)
	}

	parser.dump()

	return parser
}

func (p *Parser) readEnvFromAPIFile(filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read %v: %v", filename, err)
		return err
	}
	s := strings.Split(string(b), "=")
	if len(s) < 2 {
		return xerrors.Errorf("invalid content of %v", filename)
	}
	p.fileAPIInfo[filename] = nodeDesc{
		apiInfo: s[1],
	}

	envName := strings.TrimSpace(strings.Split(s[0], " ")[1])

	log.Infof(log.Fields{}, "set environment %v -> %v", envName, s[1])
	os.Setenv(envName, strings.TrimSpace(s[1]))

	return nil
}

func (p *Parser) parseIPFromEnvValue(val string) (string, error) {
	s := strings.Split(val, "/ip4/")
	if len(s) < 2 {
		return "", xerrors.Errorf("no valid environment available")
	}
	return strings.Split(s[1], "/tcp/")[0], nil
}

func (p *Parser) parseEnvs() {
	for key, val := range p.fileAPIInfo {
		ip, err := p.parseIPFromEnvValue(val.apiInfo)
		if err != nil {
			log.Errorf(log.Fields{}, "cannot parse %v: %v", val.apiInfo, err)
			delete(p.fileAPIInfo, key)
			continue
		}
		nd := p.fileAPIInfo[key]
		nd.ip = ip
		p.fileAPIInfo[key] = nd
	}
}

func (p *Parser) getStoragePath() {
	f, err := os.Open(CommonSetupFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot open %v: %v", CommonSetupFile, err)
		return
	}
	defer f.Close()

	bio := bufio.NewReader(f)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), " ENV_LOTUS_STORAGE_PATH") {
			s := strings.Split(string(line), "=")
			if len(s) < 2 {
				log.Errorf(log.Fields{}, "line %v do not have =", line)
				break
			}
			stoPath := s[1]
			info, err := os.Stat(stoPath)
			if err != nil {
				log.Errorf(log.Fields{}, "storage path %v: %v", stoPath, err)
				break
			}
			if !info.IsDir() {
				log.Errorf(log.Fields{}, "storage path is not dir: %v", stoPath)
				break
			}
			p.storagePath = stoPath
			p.validStoragePath = true
		}
	}
}

func (p *Parser) parseStoragePaths() {
	if !p.validStoragePath {
		return
	}
	storageCfg := filepath.Join(p.storagePath, StorageListDescFile)
	b, err := ioutil.ReadFile(storageCfg)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read %v: %v", storageCfg, err)
		return
	}
	err = json.Unmarshal(b, &p.storageConfig)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse %v: %v", storageCfg, err)
		return
	}
	p.validStorageConfig = true
}

func (p *Parser) parseLocalStorages() {
	if !p.validStorageConfig {
		return
	}
	for _, path := range p.storageConfig.StoragePaths {
		storeCfg := filepath.Join(path.Path, StorageMetaFile)
		b, err := ioutil.ReadFile(storeCfg)
		if err != nil {
			log.Errorf(log.Fields{}, "fail to read %v: %v", storeCfg, err)
			continue
		}
		meta := LocalStorageMeta{}
		err = json.Unmarshal(b, &meta)
		if err != nil {
			log.Errorf(log.Fields{}, "fail to parse meta %v: %v", storeCfg, err)
			continue
		}
		p.storageMetas = append(p.storageMetas, meta)
	}
}

func (p *Parser) getMountedCeph() {
	f, err := os.Open(ProcSelfMounts)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to open %v: %v", ProcSelfMounts, err)
		return
	}
	bio := bufio.NewReader(f)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to read %v: %v", ProcSelfMounts, err)
			break
		}
		if !strings.Contains(string(line), " ceph ") {
			continue
		}
		cephEntry := strings.Split(string(line), " ")[0]
		cephEntry = strings.Split(cephEntry, ":/")[0]
		p.cephEntries[cephEntry] = struct{}{}
	}
}

func (p *Parser) getMountedGluster() {
	log.Infof(log.Fields{}, "NOT IMPLEMENTED NOW")
}

func (p *Parser) parseMinerStorageChilds() {
	for entry := range p.cephEntries {
		s := strings.Split(entry, ",")
		for _, ss := range s {
			sss := strings.Split(ss, ":")[0]
			p.minerStorageChilds = append(p.minerStorageChilds, sss)
		}
	}
	for _, meta := range p.storageMetas {
		if meta.Oss {
			s := strings.Split(meta.OssInfo.URL, "://")[1]
			s = strings.Split(s, ":")[0]
			p.minerStorageChilds = append(p.minerStorageChilds, s)
		}
	}
}

func (p *Parser) parseLocalAddress() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		p.localAddr = strings.Split(conn.LocalAddr().String(), ":")[0]
		conn.Close()
	}
}

func (p *Parser) parseStorageHosts() {
	f, _ := os.Open(HostsFile)
	bio := bufio.NewReader(f)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to read %v: %v", HostsFile, err)
			break
		}
		s := strings.Split(string(line), " ")
		if len(s) < 2 {
			continue
		}
		if !strings.HasPrefix(s[1], "host-") {
			continue
		}
		if s[0] == p.localAddr {
			continue
		}
		p.cephStoragePeers[s[1]] = s[0]
	}
}

func (p *Parser) parseMyStorageRole() {
	resp, err := httpdaemon.R().
		Get(fmt.Sprintf("http://%v:9283", p.localAddr))
	if err == nil {
		if resp.StatusCode() == 200 {
			p.storageSubRole = types.StorageRoleMgr
			return
		}
	}

	resp, err = httpdaemon.R().
		Get(fmt.Sprintf("http://%v:9090", p.localAddr))
	if err == nil {
		if resp.StatusCode() == 200 {
			p.storageSubRole = types.StorageRoleMgr
			return
		}
	}

	resp, err = httpdaemon.R().
		Get(fmt.Sprintf("http://%v:7000", p.localAddr))
	if err == nil {
		if resp.StatusCode() == 200 {
			p.storageSubRole = types.StorageRoleAPI
			return
		}
	}

	p.storageSubRole = types.StorageRoleOsd
}

func (p *Parser) parseStorageChilds() {
	if types.StorageRoleMgr == p.storageSubRole {
		for _, v := range p.cephStoragePeers {
			p.storageChilds = append(p.storageChilds, v)
		}
	}
}

func (p *Parser) parseLogFileFromService(file string) string {
	f, _ := os.Open(file)
	bio := bufio.NewReader(f)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to read %v: %v", file, err)
			break
		}

		if !strings.HasPrefix(string(line), "Environment=GOLOG_FILE=") {
			continue
		}

		s := strings.Split(string(line), "GOLOG_FILE=")
		if len(s) < 2 {
			continue
		}

		return s[1]
	}

	return ""
}

func (p *Parser) parseLogFiles() {
	p.fullnodeLogFile = p.parseLogFileFromService(FullnodeServiceFile)
	p.minerLogFile = p.parseLogFileFromService(MinerServiceFile)
}

func (p *Parser) setEnvFromRepo(file string) {
	dir, err := p.parseRepoDirFromService(file)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse %v", err)
		return
	}
	apiPath := dir + "/api"
	b, err := ioutil.ReadFile(apiPath)
	if err != nil {
		log.Errorf(log.Fields{}, "read %v error %v", apiPath, err)
		return
	}

	apiS := strings.TrimSpace(string(b))

	dirPath := dir + "/token"
	b, err = ioutil.ReadFile(dirPath)
	if err != nil {
		log.Errorf(log.Fields{}, "read %v error %v", dirPath, err)
		return
	}
	dirS := strings.TrimSpace(string(b))

	env := dirS + ":" + apiS

	switch file {
	case MinerServiceFile:
		os.Setenv(MinerEnvKey, env)
	case FullnodeServiceFile:
		os.Setenv(FullnodeEnvKey, env)
	}
}

func (p *Parser) parseRepoDirFromService(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Errorf(log.Fields{}, "open file %v err: %v", file, err)
		return "", err
	}
	bio := bufio.NewReader(f)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			break
		}
		if !strings.HasPrefix(string(line), "ExecStart=/usr/local/bin") {
			continue
		}
		s := strings.Split(string(line), "-repo=")
		var ss string
		if len(s) > 1 {
			ss = strings.Split(s[1], " ")[0]
		}
		return ss, nil
	}
	return "", err
}

func (p *Parser) parseRepoFile() {
	minerRepoDir, err := p.parseRepoDirFromService(MinerServiceFile)
	if err != nil {
		log.Errorf(log.Fields{}, "get miner Repo dir api file err: %v", err)
	}
	p.minerRepoDirApiFile = minerRepoDir + "/api"

	fullnodeRepoDir, err := p.parseRepoDirFromService(FullnodeServiceFile)
	if err != nil {
		log.Errorf(log.Fields{}, "get fullnode Repo dir api file err: %v", err)
	}
	p.fullnodeRepoDirApiFile = fullnodeRepoDir + "/api"
}

func (p *Parser) parseApiHostFromRepoDirApiFile(file string) (string, error) {
	repoDirApiFile, err := os.Open(file)
	if err != nil {
		log.Errorf(log.Fields{}, "open file %v err: %v", file, err)
		return "", err
	}
	bio := bufio.NewReader(repoDirApiFile)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			break
		}
		if !strings.Contains(string(line), "/ip4/") {
			continue
		}
		s := strings.Split(string(line), "/")
		return s[2], nil
	}
	return "", err
}

func (p *Parser) parseApiHostFromApiFile(file string) (string, error) {
	apiFile, err := os.Open(file)
	if err != nil {
		log.Errorf(log.Fields{}, "open file %v err: %v", file, err)
		return "", err
	}
	bio := bufio.NewReader(apiFile)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			break
		}
		if !strings.Contains(string(line), "/ip4/") {
			continue
		}
		s := strings.Split(string(line), "/")
		return s[2], nil
	}
	return "", err
}

func (p *Parser) parseApiHosts() {
	fullnodeApiHost, err := p.parseApiHostFromApiFile(FullnodeAPIFile)
	if err != nil {
		fullnodeApiHost, _ = p.parseApiHostFromRepoDirApiFile(p.fullnodeRepoDirApiFile)
	}
	p.fullnodeApiHost = fullnodeApiHost

	minerApiHost, err := p.parseApiHostFromApiFile(MinerAPIFile)
	if err != nil {
		minerApiHost, _ = p.parseApiHostFromRepoDirApiFile(p.minerRepoDirApiFile)
	}
	p.minerApiHost = minerApiHost
}

func (p *Parser) parse() error {
	p.parseRepoFile()
	p.setEnvFromRepo(MinerServiceFile)
	p.setEnvFromRepo(FullnodeServiceFile)
	p.readEnvFromAPIFile(FullnodeAPIFile)
	p.readEnvFromAPIFile(MinerAPIFile)
	p.parseEnvs()
	p.getStoragePath()
	p.parseStoragePaths()
	p.parseLocalStorages()
	p.getMountedCeph()
	p.getMountedGluster()
	p.parseMinerStorageChilds()
	p.parseLocalAddress()
	p.parseStorageHosts()
	p.parseMyStorageRole()
	p.parseStorageChilds()
	p.parseLogFiles()

	p.parseApiHosts()
	return nil
}

func (p *Parser) dump() {
	fmt.Printf("PARSER ---\n")
	fmt.Printf("  API INFOS ---\n")
	for key, val := range p.fileAPIInfo {
		fmt.Printf("    %v ---\n", key)
		fmt.Printf("      env: %v\n", val.apiInfo)
		fmt.Printf("      ip:  %v\n", val.ip)
	}
	fmt.Printf("  Storage Path --- %v\n", p.storagePath)
	fmt.Printf("  Ceph Entries --\n")
	for entry := range p.cephEntries {
		fmt.Printf("    %v\n", entry)
	}
	fmt.Printf("  Ceph IPs --\n")
	for _, child := range p.minerStorageChilds {
		fmt.Printf("    %v\n", child)
	}
	fmt.Printf("  Ceph Childs ---\n")
	for k, v := range p.cephStoragePeers {
		fmt.Printf("    %v: %v\n", k, v)
	}
	fmt.Printf("  Storage Role --- %v\n", p.storageSubRole)
	fmt.Printf("  Storage Childs ---\n")
	for _, child := range p.storageChilds {
		fmt.Printf("    %v\n", child)
	}
}

func (p *Parser) GetParentIP(myRole string) (string, error) {
	switch myRole {
	case types.FullNode:
		return "", xerrors.Errorf("fullnode do not have parent")
	case types.MinerNode:
		if _, ok := p.fileAPIInfo[FullnodeAPIFile]; !ok {
			return "", xerrors.Errorf("do not have miner api info")
		}
		return p.fileAPIInfo[FullnodeAPIFile].ip, nil
	case types.FullMinerNode:
		return "", xerrors.Errorf("fullminernode do not have parent")
	case types.WorkerNode:
		if _, ok := p.fileAPIInfo[MinerAPIFile]; !ok {
			return "", xerrors.Errorf("do not have miner api info")
		}
		return p.fileAPIInfo[MinerAPIFile].ip, nil
	case types.StorageNode:
		return "", xerrors.Errorf("storagenode do not have parent")
	}
	return "", xerrors.Errorf("unknow role %v", myRole)
}

func (p *Parser) getMinerStorageChilds() ([]string, error) {
	return p.minerStorageChilds, nil
}

func (p *Parser) getStorageChilds() ([]string, error) {
	return p.storageChilds, nil
}

func (p *Parser) GetChildsIPs(myRole string) ([]string, error) {
	switch myRole {
	case types.FullMinerNode:
		return p.getMinerStorageChilds()
	case types.MinerNode:
		return p.getMinerStorageChilds()
	case types.StorageNode:
		return p.getStorageChilds()
	}
	return nil, xerrors.Errorf("no child for %v", myRole)
}

func (p *Parser) GetSubRole(myRole string) (string, error) {
	switch myRole {
	case types.StorageNode:
		return p.storageSubRole, nil
	}
	return "", xerrors.Errorf("no subrole for %v", myRole)
}

func (p *Parser) GetLogFile(myRole string) (string, error) {
	switch myRole {
	case types.MinerNode:
		return p.minerLogFile, nil
	case types.FullNode:
		return p.fullnodeLogFile, nil
	case types.ChiaMinerNode:
		return p.chiaMinerNodeLogFile, nil
	case types.ChiaPlotterNode:
		return p.chiaPlotterLogFile, nil
	default:
		return "", xerrors.Errorf("no log file for role: %v", myRole)
	}
}

func (p *Parser) GetShareStorageRoot(myRole string) (string, error) {
	switch myRole {
	case types.MinerNode:
		fallthrough
	case types.FullMinerNode:
		return p.minerShareStorageRoot, nil
	default:
		return "", xerrors.Errorf("no share storage for role: %v", myRole)
	}
}

func (p *Parser) GetApiHostByHostRole(myRole string) (string, error) {
	switch myRole {
	case types.MinerNode:
		return p.minerApiHost, nil
	case types.FullNode:
		return p.fullnodeApiHost, nil
	default:
		return "", xerrors.Errorf("no api host for role: %v", myRole)
	}

}
