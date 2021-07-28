package miner

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/NpoolDevOps/fbc-devops-peer/metrics/minermetrics"
	"github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/prometheus/client_golang/prometheus"
)

type MinerNode struct {
	*basenode.Basenode
	minerMetrics *minermetrics.MinerMetrics
}

func NewMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *MinerNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	miner := &MinerNode{
		basenode.NewBasenode(config, devopsClient),
		nil,
	}

	logfile, _ := miner.GetLogFileByRole(types.MinerNode)
	shareStorageRoot, _ := miner.GetShareStorageRootByRole(types.MinerNode)
	miner.minerMetrics = minermetrics.NewMinerMetrics(minermetrics.MinerMetricsConfig{
		ShareStorageRoot: shareStorageRoot,
		Logfile:          logfile,
	})

	miner.SetAddrNotifier(miner.addressNotifier)
	return miner
}

func (n *MinerNode) addressNotifier(local, public string) {
	n.minerMetrics.SetHost(local)
	fullnodeHost, _ := n.GetFullnodeHost()
	n.minerMetrics.SetFullnodeHost(fullnodeHost)
}

func (n *MinerNode) Describe(ch chan<- *prometheus.Desc) {
	n.minerMetrics.Describe(ch)
	n.BaseMetrics.Describe(ch)
}

func (n *MinerNode) Collect(ch chan<- prometheus.Metric) {
	n.minerMetrics.Collect(ch)
	n.BaseMetrics.Collect(ch)
}

func (n *MinerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}

func (n *MinerNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      MMMMMIIIIINNNNEEEERRRR      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
