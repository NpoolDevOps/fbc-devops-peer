package workermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/prometheus/client_golang/prometheus"
)

type WorkerMetrics struct {
	OpenFileNumber *prometheus.Desc
}

func NewWorkerMetrics() *WorkerMetrics {
	metrics := &WorkerMetrics{
		OpenFileNumber: prometheus.NewDesc(
			"worker_open_file_number",
			"show worker open file number",
			nil, nil,
		),
	}
	return metrics
}

func (m *WorkerMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.OpenFileNumber
}

func (w *WorkerMetrics) Collect(ch chan<- prometheus.Metric) {
	workerOpenFileNumber, err := systemapi.GetProcessOpenFileNumber("lotus-worker")
	if err == nil {
		ch <- prometheus.MustNewConstMetric(w.OpenFileNumber, prometheus.CounterValue, float64(workerOpenFileNumber))
	}
}
