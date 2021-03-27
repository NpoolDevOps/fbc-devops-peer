package fullnode

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
)

type FullNode struct {
	*basenode.Basenode
}

func NewFullNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	fullnode := &FullNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return fullnode
}
