package worker

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
)

type WorkerNode struct {
	*basenode.Basenode
}

func NewWorkerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *WorkerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	worker := &WorkerNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return worker
}
