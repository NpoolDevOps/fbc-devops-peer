package lotusmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type LotusMetrics struct {
	HeightDiff   *prometheus.Desc
	BlockElapsed *prometheus.Desc
	NetPeers     *prometheus.Desc
}

func NewLotusMetrics() *LotusMetrics {
	return &LotusMetrics{
		HeightDiff: prometheus.NewDesc(
			"lotus_chain_height_diff",
			"Show lotus chain sync height diff",
			nil, nil,
		),
		BlockElapsed: prometheus.NewDesc(
			"lotus_chain_block_elapsed",
			"Show lotus chain elapsed time of current block height",
			nil, nil,
		),
		NetPeers: prometheus.NewDesc(
			"lotus_client_net_peers",
			"Show how many peers are connected by lotus client",
			nil, nil,
		),
	}
}
