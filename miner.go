package main

import (
	log "github.com/EntropyPool/entropy-logger"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Miner struct {
	basenode *Basenode
}

func parseIPFromEnvironment(env string) (string, error) {
	s := strings.Split(env, "/ip4/")
	if len(s) < 2 {
		return "", xerrors.Errorf("no valid environment available")
	}
	return strings.Split(s[1], "/tcp/")[0], nil
}

func parseParentIP() (string, error) {
	profile := "/etc/profile.d/fullnode-api-info.sh"
	b, err := ioutil.ReadFile(profile)
	if err == nil {
		ip, err := parseIPFromEnvironment(string(b))
		if err == nil {
			return ip, nil
		}
	}

	env, ok := os.LookupEnv("FULLNODE_API_INFO")
	if !ok {
		return "", xerrors.Errorf("no valid environment available")
	}
	return parseIPFromEnvironment(env)
}

func NewMinerPeer(config *BasenodeConfig) *Miner {
	miner := &Miner{}

	miner.basenode = NewBasenode(config)
	if miner.basenode == nil {
		log.Errorf(log.Fields{}, "fail to create devops client")
		return nil
	}

	miner.basenode.ReportParentSpec(parseParentIP)

	return miner
}

func (n *Miner) Run() error {
	ticker := time.NewTicker(3 * time.Minute)
	for {
		<-ticker.C
	}
}
