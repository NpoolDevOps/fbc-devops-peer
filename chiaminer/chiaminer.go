package chiaminer

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	chiaMinerMetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/chiaminermetrics"
	"github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaMinerNode struct {
	*basenode.Basenode
	chiaMinerMetrics *chiaMinerMetrics.ChiaMinerMetrics
}

func NewChiaMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *ChiaMinerNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	chiaminer := &ChiaMinerNode{
		basenode.NewBasenode(config, devopsClient),
		nil,
	}

	logfile, _ := chiaminer.GetLogFileByRole(types.ChiaMinerNode)
	chiaminer.chiaMinerMetrics = chiaMinerMetrics.NewChiaMinerMetrics(logfile)

	chiaminer.SetAddrNotifier(chiaminer.addressNotifier)
	return chiaminer
}

func (c *ChiaMinerNode) addressNotifier(local, public string) {
	c.chiaMinerMetrics.SetHost(local)
	c.chiaMinerMetrics.SetFullnodeHost(local)
}

func (c *ChiaMinerNode) Describe(ch chan<- *prometheus.Desc) {
	c.chiaMinerMetrics.Describe(ch)
	c.BaseMetrics.Describe(ch)
}

func (c *ChiaMinerNode) Collect(ch chan<- prometheus.Metric) {
	c.chiaMinerMetrics.Collect(ch)
	c.BaseMetrics.Collect(ch)
}

func (c *ChiaMinerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(c)
}

func (c *ChiaMinerNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      CCHHIIAAAMMMIIINNNEERR       ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
