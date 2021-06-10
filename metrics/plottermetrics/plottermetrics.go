package plottermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/chiaapi"
	api "github.com/NpoolDevOps/fbc-devops-peer/api/plotterapi"
	"github.com/prometheus/client_golang/prometheus"
)

type PlotterMetrics struct {
	PlotterTime         *prometheus.Desc //p盘时间
	PlotterProcess      *prometheus.Desc //进程数
	PlotterStatus       *prometheus.Desc //p盘状态
	StorageProxyStatus  *prometheus.Desc //chia-storage-proxy状态
	StorageServerStatus *prometheus.Desc
	PlotterTimeCount    *prometheus.Desc

	host    string
	hasHost bool
}

func NewPlotterMetrics(logfile string) *PlotterMetrics {
	pp := &PlotterMetrics{
		PlotterTime: prometheus.NewDesc(
			"plotter_time",
			"show plotter average time",
			nil, nil,
		),
		PlotterStatus: prometheus.NewDesc(
			"plotter_status",
			"show plotter status",
			nil, nil,
		),
		PlotterProcess: prometheus.NewDesc(
			"plotter_process",
			"show plotter process",
			nil, nil,
		),
		StorageProxyStatus: prometheus.NewDesc(
			"chia_storage_proxy_status",
			"show chia storage proxy status",
			nil, nil,
		),
		PlotterTimeCount: prometheus.NewDesc(
			"plotter_time_count",
			"show plotter plot times count",
			nil, nil,
		),
		StorageServerStatus: prometheus.NewDesc(
			"storage_server_status",
			"show chia storage server status",
			nil, nil,
		),
	}
	return pp
}

func (p *PlotterMetrics) SetHost(host string) {
	p.host = host
	p.hasHost = true
}

func (p *PlotterMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.PlotterProcess
	ch <- p.PlotterTime
	ch <- p.PlotterStatus
	ch <- p.StorageProxyStatus
	ch <- p.PlotterTimeCount
	ch <- p.StorageServerStatus
}

func (p *PlotterMetrics) Collect(ch chan<- prometheus.Metric) {
	plotterProcessNum, _ := api.GetPlotterProcessNum()
	plotterTime, plotterTimeCount, _ := api.GetPlotterTime()
	plotterStatus, _ := chiaapi.GetStatus("ProofOfSpace create")
	storageProxyStatus, _ := chiaapi.GetStatus("chia-storage-proxy")
	storageServerStatus, _ := chiaapi.GetStatus("chia-storage-server")
	ps := 0
	sps := 0
	sss := 0
	if plotterStatus == "active" {
		ps = 1
	}
	if storageProxyStatus == "active" {
		sps = 1
	}
	if storageServerStatus == "active" {
		sss = 1
	}

	ch <- prometheus.MustNewConstMetric(p.PlotterProcess, prometheus.CounterValue, float64(plotterProcessNum))
	ch <- prometheus.MustNewConstMetric(p.PlotterStatus, prometheus.CounterValue, float64(ps))
	ch <- prometheus.MustNewConstMetric(p.StorageProxyStatus, prometheus.CounterValue, float64(sps))
	ch <- prometheus.MustNewConstMetric(p.PlotterTime, prometheus.CounterValue, float64(plotterTime))
	ch <- prometheus.MustNewConstMetric(p.PlotterTimeCount, prometheus.CounterValue, float64(plotterTimeCount))
	ch <- prometheus.MustNewConstMetric(p.StorageServerStatus, prometheus.CounterValue, float64(sss))
}
