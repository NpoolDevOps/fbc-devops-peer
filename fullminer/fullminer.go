package fullminer

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
)

type FullMinerNode struct {
	*basenode.Basenode
}

func NewFullMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullMinerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	fullminer := &FullMinerNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return fullminer
}
