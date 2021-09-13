package worker

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/NpoolDevOps/fbc-devops-peer/metrics/workermetrics"
	"github.com/prometheus/client_golang/prometheus"
)

type WorkerNode struct {
	*basenode.Basenode
	workermetrics *workermetrics.WorkerMetrics
}

func NewWorkerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *WorkerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	worker := &WorkerNode{
		basenode.NewBasenode(config, devopsClient),
		nil,
	}
	worker.workermetrics = workermetrics.NewWorkerMetrics(worker.Username, worker.NetworkType)
	return worker
}

func (n *WorkerNode) Describe(ch chan<- *prometheus.Desc) {
	n.BaseMetrics.Describe(ch)
	n.workermetrics.Describe(ch)
}

func (n *WorkerNode) Collect(ch chan<- prometheus.Metric) {
	n.BaseMetrics.Collect(ch)
	n.workermetrics.Collect(ch)
}

func (n *WorkerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}

func (n *WorkerNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      WWWWWWOOOOORRRRRKKKKKKK      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
