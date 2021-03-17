package main

import (
	log "github.com/EntropyPool/entropy-logger"
)

type Miner struct {
	basenode *Basenode
}

func NewMinerPeer(config *PeerConfig) *Miner {
	miner := &Miner{}

	miner.basenode = NewBasenode(&BasenodeConfig{
		DevopsConfig: &DevopsConfig{
			PeerConfig: config,
		},
	})
	if miner.basenode == nil {
		log.Errorf(log.Fields{}, "fail to create devops client")
		return nil
	}
	return miner
}

func (n *Miner) Run() error {
	return nil
}
