package plotter

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	"github.com/NpoolDevOps/fbc-devops-peer/devops"
	"github.com/NpoolDevOps/fbc-devops-peer/exporter"
	plotterMetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/plottermetrics"
	"github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/prometheus/client_golang/prometheus"
)

type PlotterNode struct {
	*basenode.Basenode
	plotterMetrics *plotterMetrics.PlotterMetrics
}

func NewPlotterNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *PlotterNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	plotter := &PlotterNode{
		basenode.NewBasenode(config, devopsClient),
		nil,
	}

	logfile, _ := plotter.GetLogFileByRole(types.PlotterNode)
	plotter.plotterMetrics = plotterMetrics.NewPlotterMetrics(logfile)

	plotter.SetAddrNotifier(plotter.addressNotifier)
	return plotter
}

func (p *PlotterNode) addressNotifier(local, public string) {
	p.plotterMetrics.SetHost(local)
	p.plotterMetrics.SetFullnodeHost(local)
}

func (p *PlotterNode) Describe(ch chan<- *prometheus.Desc) {
	p.plotterMetrics.Describe(ch)
	p.BaseMetrics.Describe(ch)
}

func (p *PlotterNode) Collect(ch chan<- prometheus.Metric) {
	p.plotterMetrics.Collect(ch)
	p.BaseMetrics.Collect(ch)
}

func (p *PlotterNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(p)
}

func (p *PlotterNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      PPPPLLLLOOOOOTTTTTEERR       ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
