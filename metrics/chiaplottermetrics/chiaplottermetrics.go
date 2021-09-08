package chiaplottermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/chiaplotterlog"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaPlotterMetrics struct {
	cpl                       *chiaplotterlog.ChiaPlotterLog
	PlotterAvgTime            *prometheus.Desc
	PlotterMaxTime            *prometheus.Desc
	PlotterMinTime            *prometheus.Desc
	PlotterPlotCount          *prometheus.Desc
	PlotterProcessCount       *prometheus.Desc
	StorageProxyProcessCount  *prometheus.Desc
	StorageServerProcessCount *prometheus.Desc

	host     string
	hasHost  bool
	username string
}

func NewChiaPlotterMetrics(logfile, username string) *ChiaPlotterMetrics {
	cpm := &ChiaPlotterMetrics{
		username: username,
		PlotterAvgTime: prometheus.NewDesc(
			"plotter_average_time",
			"show plotter average time",
			[]string{"user"}, nil,
		),
		PlotterMaxTime: prometheus.NewDesc(
			"plotter_max_time",
			"show the max value of plotter time",
			[]string{"user"}, nil,
		),
		PlotterMinTime: prometheus.NewDesc(
			"plotter_min_time",
			"show the min value of plotter time",
			[]string{"user"}, nil,
		),
		PlotterPlotCount: prometheus.NewDesc(
			"plotter_plot_count",
			"show times parse plotter time",
			[]string{"user"}, nil,
		),
		PlotterProcessCount: prometheus.NewDesc(
			"plotter_process_count",
			"show plotter status",
			[]string{"user"}, nil,
		),
		StorageProxyProcessCount: prometheus.NewDesc(
			"chia_storage_proxy_process_count",
			"show chia storage proxy process count",
			[]string{"user"}, nil,
		),
		StorageServerProcessCount: prometheus.NewDesc(
			"storage_server_process_count",
			"show chia storage server process count",
			[]string{"user"}, nil,
		),
	}
	return cpm
}

func (p *ChiaPlotterMetrics) SetHost(host string) {
	p.host = host
	p.hasHost = true
}

func (p *ChiaPlotterMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.PlotterAvgTime
	ch <- p.PlotterMaxTime
	ch <- p.PlotterMinTime
	ch <- p.PlotterPlotCount
	ch <- p.PlotterProcessCount
	ch <- p.StorageProxyProcessCount
	ch <- p.StorageServerProcessCount
}

func (p *ChiaPlotterMetrics) Collect(ch chan<- prometheus.Metric) {
	username := p.username

	plotterAvgTime := p.cpl.GetChiaPlotterAvgTime()
	plotterMaxTime := p.cpl.GetChiaPlotterMaxTime()
	plotterMinTime := p.cpl.GetChiaPlotterMinTime()
	PlotterPlotCount := p.cpl.GetParseChiaPlotterTimeCount()
	plotterProcessCount, _ := systemapi.GetProcessCount("/usr/local/bin/chia_plot -2")
	storageProxyStatus, _ := systemapi.GetProcessCount("chia-storage-proxy")
	storageServerStatus, _ := systemapi.GetProcessCount("chia-storage-server")

	ch <- prometheus.MustNewConstMetric(p.PlotterProcessCount, prometheus.CounterValue, float64(plotterProcessCount), username)
	ch <- prometheus.MustNewConstMetric(p.StorageProxyProcessCount, prometheus.CounterValue, float64(storageProxyStatus), username)
	ch <- prometheus.MustNewConstMetric(p.StorageServerProcessCount, prometheus.CounterValue, float64(storageServerStatus), username)
	ch <- prometheus.MustNewConstMetric(p.PlotterAvgTime, prometheus.CounterValue, plotterAvgTime, username)
	ch <- prometheus.MustNewConstMetric(p.PlotterMaxTime, prometheus.CounterValue, plotterMaxTime, username)
	ch <- prometheus.MustNewConstMetric(p.PlotterMinTime, prometheus.CounterValue, plotterMinTime, username)
	ch <- prometheus.MustNewConstMetric(p.PlotterPlotCount, prometheus.CounterValue, float64(PlotterPlotCount), username)
}
