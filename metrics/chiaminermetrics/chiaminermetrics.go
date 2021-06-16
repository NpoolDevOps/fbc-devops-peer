package chiaminermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/chiaminerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaMinerMetrics struct {
	cml                   *chiaminerlog.ChiaMinerLog
	ChiaMinerProcessCount *prometheus.Desc //chia miner服务状态
	ChiaMinerTimeoutCount *prometheus.Desc //chia miner 扫盘超时

	host    string
	hasHost bool
}

func NewChiaMinerMetrics(logfile string) *ChiaMinerMetrics {
	cmm := &ChiaMinerMetrics{
		cml: chiaminerlog.NewChiaMinerLog(logfile),
		ChiaMinerProcessCount: prometheus.NewDesc(
			"chia_miner_process_count",
			"show chia miner process count",
			nil, nil,
		),
		ChiaMinerTimeoutCount: prometheus.NewDesc(
			"chia_miner_timeout_count",
			"show chia miner timeout count",
			nil, nil,
		),
	}
	return cmm
}

func (c *ChiaMinerMetrics) SetHost(host string) {
	c.host = host
	c.hasHost = true
}

func (c *ChiaMinerMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ChiaMinerProcessCount
	ch <- c.ChiaMinerTimeoutCount
}

func (c *ChiaMinerMetrics) Collect(ch chan<- prometheus.Metric) {
	chiaMinerProcessCount, _ := systemapi.GetProcessCount("/usr/local/bin/hpool-miner-chia -config")
	chiaMinerTimeout := c.cml.GetChiaMinerTimeout()

	ch <- prometheus.MustNewConstMetric(c.ChiaMinerProcessCount, prometheus.CounterValue, float64(chiaMinerProcessCount))
	ch <- prometheus.MustNewConstMetric(c.ChiaMinerTimeoutCount, prometheus.CounterValue, float64(chiaMinerTimeout))

}
