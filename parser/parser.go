package parser

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	"golang.org/x/xerrors"
	"io/ioutil"
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
)

type nodeDesc struct {
	apiInfo string
	ip      string
}

type Parser struct {
	fileAPIInfo map[string]nodeDesc
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

func (p *Parser) parseAPIInfo(info string) error {
	return nil
}

func (p *Parser) readEnvFromAPIFile(filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
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
			delete(p.fileAPIInfo, key)
			continue
		}
		nd := p.fileAPIInfo[key]
		nd.ip = ip
		p.fileAPIInfo[key] = nd
	}
}

func (p *Parser) parse() error {
	p.readEnvFromAPIFile(FullnodeAPIFile)
	p.readEnvFromAPIFile(MinerAPIFile)
	p.parseEnvs()
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
