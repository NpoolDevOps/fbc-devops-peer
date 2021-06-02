package storage

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/prometheus/client_golang/prometheus"
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

func (n *StorageNode) Describe(ch chan<- *prometheus.Desc) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR STORAGENODE")
}

func (n *StorageNode) Collect(ch chan<- prometheus.Metric) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR STORAGENODE")
}

func (n *StorageNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}
