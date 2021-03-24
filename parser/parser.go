package parser

import (
	log "github.com/EntropyPool/entropy-logger"
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
