package storage

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
)

type StorageNode struct {
	*basenode.Basenode
}

func NewStorageNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *StorageNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	storage := &StorageNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return storage
}
