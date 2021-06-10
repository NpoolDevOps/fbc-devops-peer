package plottermetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PlotterMetrics struct {
	PlotTime           *prometheus.Desc //p盘时间
	PlotProcess        *prometheus.Desc //进程数
	PlotterStatus      *prometheus.Desc //p盘状态
	StorageProxyStatus *prometheus.Desc //chia-storage-proxy状态

	host    string
	hasHost bool
}

func NewPlotterMetrics(logfile string) *PlotterMetrics {
	return nil
}

func (p *PlotterMetrics) SetHost(host string) {
	p.host = host
	p.hasHost = true
}

func (p *PlotterMetrics) SetFullnodeHost(host string) {

}

func (p *PlotterMetrics) Describe(ch chan<- *prometheus.Desc) {}

func (p *PlotterMetrics) Collect(ch chan<- prometheus.Metric) {}
