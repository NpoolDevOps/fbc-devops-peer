package chiaminermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/chiaminerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaMinerMetrics struct {
	cml                  *chiaminerlog.ChiaMinerLog
	ChiaMinerStatusError *prometheus.Desc //chia miner服务状态
	ChiaMinerTimeout     *prometheus.Desc //chia miner 扫盘超时

	host    string
	hasHost bool
}

func NewChiaMinerMetrics(logfile string) *ChiaMinerMetrics {
	cmm := &ChiaMinerMetrics{
		cml: chiaminerlog.NewChiaMinerLog(logfile),
		ChiaMinerStatusError: prometheus.NewDesc(
			"chia_miner_status_error",
			"show chia miner status",
			nil, nil,
		),
		ChiaMinerTimeout: prometheus.NewDesc(
			"chia_miner_timeout",
			"show chia miner timeout",
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
	ch <- c.ChiaMinerStatusError
	ch <- c.ChiaMinerTimeout
}

func (c *ChiaMinerMetrics) Collect(ch chan<- prometheus.Metric) {
	chiaMinerStatusError, _ := systemapi.GetProcessStatusError("/usr/local/bin/hpool-miner-chia -config")
	chiaMinerTimeout := c.cml.GetChiaMinerTimeout()

	ch <- prometheus.MustNewConstMetric(c.ChiaMinerStatusError, prometheus.CounterValue, float64(chiaMinerStatusError))
	if chiaMinerTimeout {
		ch <- prometheus.MustNewConstMetric(c.ChiaMinerTimeout, prometheus.CounterValue, float64(1))
	} else {
		ch <- prometheus.MustNewConstMetric(c.ChiaMinerTimeout, prometheus.CounterValue, float64(0))
	}

}
