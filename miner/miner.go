package miner

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
)

type MinerNode struct {
	*basenode.Basenode
}

func NewMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *MinerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	miner := &MinerNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return miner
}
