package chiaplottermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/chiaplotterlog"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaPlotterMetrics struct {
	cpl                       *chiaplotterlog.ChiaPlotterLog
	PlotterPlottingTime       *prometheus.Desc
	PlotterProcessCount       *prometheus.Desc
	StorageProxyProcessCount  *prometheus.Desc
	StorageServerProcessCount *prometheus.Desc

	host    string
	hasHost bool
}

func NewChiaPlotterMetrics(logfile string) *ChiaPlotterMetrics {
	cpm := &ChiaPlotterMetrics{
		PlotterPlottingTime: prometheus.NewDesc(
			"plotter_plotting_time",
			"show plotter average time",
			nil, nil,
		),
		PlotterProcessCount: prometheus.NewDesc(
			"plotter_Process_count",
			"show plotter status",
			nil, nil,
		),
		StorageProxyProcessCount: prometheus.NewDesc(
			"chia_storage_proxy_process_count",
			"show chia storage proxy process count",
			nil, nil,
		),
		StorageServerProcessCount: prometheus.NewDesc(
			"storage_server_process_count",
			"show chia storage server process count",
			nil, nil,
		),
	}
	return cpm
}

func (p *ChiaPlotterMetrics) SetHost(host string) {
	p.host = host
	p.hasHost = true
}

func (p *ChiaPlotterMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.PlotterPlottingTime
	ch <- p.PlotterProcessCount
	ch <- p.StorageProxyProcessCount
	ch <- p.StorageServerProcessCount
}

func (p *ChiaPlotterMetrics) Collect(ch chan<- prometheus.Metric) {
	plottingTime := p.cpl.GetChiaPlotterTime()
	plotterProcessCount, _ := systemapi.GetProcessCount("/usr/local/bin/chia_plot -2")
	storageProxyStatus, _ := systemapi.GetProcessCount("chia-storage-proxy")
	storageServerStatus, _ := systemapi.GetProcessCount("chia-storage-server")

	ch <- prometheus.MustNewConstMetric(p.PlotterProcessCount, prometheus.CounterValue, float64(plotterProcessCount))
	ch <- prometheus.MustNewConstMetric(p.StorageProxyProcessCount, prometheus.CounterValue, float64(storageProxyStatus))
	ch <- prometheus.MustNewConstMetric(p.PlotterPlottingTime, prometheus.CounterValue, plottingTime)
	ch <- prometheus.MustNewConstMetric(p.StorageServerProcessCount, prometheus.CounterValue, float64(storageServerStatus))
}
