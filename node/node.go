package node

import (
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type Node interface {
	GetMainRole() string
	GetSubRole() string
	NotifyParentSpec(string)
	GetParentIP() (string, error)
	GetChildsIPs() ([]string, error)
	NotifyPeerId(uuid.UUID)
	Banner()
	SetPeer(interface{})
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
	CreateExporter() *exporter.Exporter
}
