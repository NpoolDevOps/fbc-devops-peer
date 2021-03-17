package main

import (
	log "github.com/EntropyPool/entropy-logger"
	"time"
)

type Miner struct {
	basenode *Basenode
}

func NewMinerPeer(config *PeerConfig) *Miner {
	miner := &Miner{}

	miner.basenode = NewBasenode(&BasenodeConfig{
		PeerConfig: config,
	})
	if miner.basenode == nil {
		log.Errorf(log.Fields{}, "fail to create devops client")
		return nil
	}
	return miner
}

func (n *Miner) Run() error {
	ticker := time.NewTicker(3 * time.Minute)
	for {
		<-ticker.C
	}
}
