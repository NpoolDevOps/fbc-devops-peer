package chiaplotter

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	"github.com/NpoolDevOps/fbc-devops-peer/devops"
	"github.com/NpoolDevOps/fbc-devops-peer/exporter"
	metrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/chiaplottermetrics"
	"github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaPlotterNode struct {
	*basenode.Basenode
	plotterMetrics *metrics.ChiaPlotterMetrics
}

func NewPlotterNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *ChiaPlotterNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	plotter := &ChiaPlotterNode{
		basenode.NewBasenode(config, devopsClient),
		nil,
	}

	logfile, _ := plotter.GetLogFileByRole(types.ChiaPlotterNode)
	plotter.plotterMetrics = metrics.NewChiaPlotterMetrics(logfile, plotter.Username)

	plotter.SetAddrNotifier(plotter.addressNotifier)
	return plotter
}

func (p *ChiaPlotterNode) addressNotifier(local, public string) {
	p.plotterMetrics.SetHost(local)
}

func (p *ChiaPlotterNode) Describe(ch chan<- *prometheus.Desc) {
	p.plotterMetrics.Describe(ch)
	p.BaseMetrics.Describe(ch)
}

func (p *ChiaPlotterNode) Collect(ch chan<- prometheus.Metric) {
	p.plotterMetrics.Collect(ch)
	p.BaseMetrics.Collect(ch)
}

func (p *ChiaPlotterNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(p)
}

func (p *ChiaPlotterNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      PPPPLLLLOOOOOTTTTTEERR       ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
