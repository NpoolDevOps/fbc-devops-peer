package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
)

type nodeDesc struct {
	apiInfo string
	ip      string
}

type storageDesc struct {
	storageType string
	ips         []string
	vendor      string
}

type Parser struct {
	fileAPIInfo        map[string]nodeDesc
	storageChilds      []storageDesc
	storagePath        string
	validStoragePath   bool
	storageConfig      StorageConfig
	validStorageConfig bool
	storageMetas       []LocalStorageMeta
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
		fileAPIInfo: map[string]nodeDesc{},
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
		if strings.Contains(string(line), "ENV_LOTUS_STORAGE_PATH") {
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
				log.Errorf(log.Fields{}, "storage path is not dir", stoPath)
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

	}
}

func (p *Parser) parse() error {
	p.readEnvFromAPIFile(FullnodeAPIFile)
	p.readEnvFromAPIFile(MinerAPIFile)
	p.parseEnvs()
	p.getStoragePath()
	p.parseStoragePaths()
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
	return nil, nil
}

func (p *Parser) GetChildIPs(myRole string) ([]string, error) {
	switch myRole {
	case types.MinerNode:
		return p.getMinerStorageChilds()
	}
	return nil, xerrors.Errorf("no child for %v", myRole)
}
