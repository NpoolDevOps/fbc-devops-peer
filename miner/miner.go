package miner

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/NpoolDevOps/fbc-devops-peer/metrics/minermetrics"
	"github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/NpoolDevOps/fbc-devops-peer/version"
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
	miner.minerMetrics = minermetrics.NewMinerMetrics(logfile)

	miner.SetAddrNotifier(miner.addressNotifier)
	miner.WatchVersions(miner.getVersions)
	return miner
}

func (n *MinerNode) addressNotifier(local, public string) {
	n.minerMetrics.SetHost(local)
	n.minerMetrics.SetFullnodeHost(local)
}

func (n *MinerNode) getVersions(host string) []version.Version {
	vers := []version.Version{}

	ver, err := lotusapi.ClientVersion(host)
	if err == nil {
		vers = append(vers, ver)
	}

	return vers
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
