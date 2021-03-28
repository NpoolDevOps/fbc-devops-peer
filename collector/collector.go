package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Collector interface {
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
}
