package chiaminermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/chiaapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/chiaminerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type ChiaMinerMetrics struct {
	cml              *chiaminerlog.ChiaMinerLog
	ChiaMinerStatus  *prometheus.Desc //chia miner服务状态
	ChiaMinerTimeout *prometheus.Desc //chia miner 扫盘超时

	host    string
	hasHost bool
}

func NewChiaMinerMetrics(logfile string) *ChiaMinerMetrics {
	cmm := &ChiaMinerMetrics{
		cml: chiaminerlog.NewChiaMinerLog(logfile),
		ChiaMinerStatus: prometheus.NewDesc(
			"chia_miner_status",
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
	ch <- c.ChiaMinerStatus
	ch <- c.ChiaMinerTimeout
}

func (c *ChiaMinerMetrics) Collect(ch chan<- prometheus.Metric) {
	chiaMinerStatus, _ := chiaapi.GetStatus("/usr/local/bin/hpool-miner-chia -config")
	chiaMinerTimeout := c.cml.GetChiaMinerTimeout()

	cmt := 0
	cms := 0
	if chiaMinerStatus == "active" {
		cms = 1
	}
	if chiaMinerTimeout {
		cmt = 1
	}

	ch <- prometheus.MustNewConstMetric(c.ChiaMinerStatus, prometheus.CounterValue, float64(cms))
	ch <- prometheus.MustNewConstMetric(c.ChiaMinerTimeout, prometheus.CounterValue, float64(cmt))
}
