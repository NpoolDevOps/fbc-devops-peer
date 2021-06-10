package chiaminermetrics

import "github.com/prometheus/client_golang/prometheus"

type ChiaMinerMetrics struct {
	host    string
	hasHost bool
}

func NewChiaMinerMetrics(logfile string) *ChiaMinerMetrics {
	return nil
}

func (c *ChiaMinerMetrics) SetHost(host string) {
	c.host = host
	c.hasHost = true
}

func (c *ChiaMinerMetrics) SetFullnodeHost(host string) {
	// c.fullnodeHost = host
	// c.ml.SetFullnodeHost(host)
}

func (c *ChiaMinerMetrics) Describe(ch chan<- *prometheus.Desc) {}

func (c *ChiaMinerMetrics) Collect(ch chan<- prometheus.Metric) {}
