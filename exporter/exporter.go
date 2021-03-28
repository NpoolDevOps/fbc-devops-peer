package exporter

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	collector "github.com/NpoolDevOps/fbc-devops-peer/collector"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type Exporter struct {
}

func NewExporter(collector collector.Collector) *Exporter {
	prometheus.MustRegister(collector)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		for {
			listen := fmt.Sprintf(":%v", types.ExporterPort)
			log.Infof(log.Fields{}, "Run exporter at %v", listen)
			http.ListenAndServe(listen, nil)
			time.Sleep(1 * time.Minute)
		}
	}()
	return &Exporter{}
}
