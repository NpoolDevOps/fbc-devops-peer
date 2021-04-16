package minermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/minerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type MinerMetrics struct {
	ml             *minerlog.MinerLog
	ForkBlocks     *prometheus.Desc
	PastBlocks     *prometheus.Desc
	FailedBlocks   *prometheus.Desc
	BlockTookAvgMs *prometheus.Desc
	BlockTookMaxMs *prometheus.Desc
	BlockTookMinMs *prometheus.Desc
	Blocks         *prometheus.Desc

	SectorTaskElapsed    *prometheus.Desc
	SectorTaskDuration   *prometheus.Desc
	SectorTaskConcurrent *prometheus.Desc
	SectorTaskDones      *prometheus.Desc
	SectorTaskProgress   *prometheus.Desc

	MinerSectorTaskConcurrent *prometheus.Desc
	MinerSectorTaskDones      *prometheus.Desc

	errors       int
	host         string
	hasHost      bool
	fullnodeHost string
}

func NewMinerMetrics(logfile string) *MinerMetrics {
	mm := &MinerMetrics{
		ml: minerlog.NewMinerLog(logfile),
		ForkBlocks: prometheus.NewDesc(
			"miner_fork_blocks",
			"Show miner fork blocks",
			nil, nil,
		),
		PastBlocks: prometheus.NewDesc(
			"miner_block_in_past",
			"Show miner block in past",
			nil, nil,
		),
		FailedBlocks: prometheus.NewDesc(
			"miner_block_failed",
			"Show miner block failed",
			nil, nil,
		),
		BlockTookAvgMs: prometheus.NewDesc(
			"miner_block_took_average_ms",
			"Show miner block took average ms",
			nil, nil,
		),
		BlockTookMaxMs: prometheus.NewDesc(
			"miner_block_took_max_ms",
			"Show miner block took max ms",
			nil, nil,
		),
		BlockTookMinMs: prometheus.NewDesc(
			"miner_block_took_min_ms",
			"Show miner block took min ms",
			nil, nil,
		),
		Blocks: prometheus.NewDesc(
			"miner_block_produced",
			"Show miner block produced",
			nil, nil,
		),
		SectorTaskElapsed: prometheus.NewDesc(
			"miner_seal_sector_task_elapsed",
			"Miner seal sector task elapsed",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskDuration: prometheus.NewDesc(
			"miner_seal_sector_task_duration",
			"Miner seal sector task duration",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskConcurrent: prometheus.NewDesc(
			"miner_seal_sector_task_concurrent",
			"Miner seal sector task concurrent",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskDones: prometheus.NewDesc(
			"miner_seal_sector_task_dones",
			"Miner seal sector task dones",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskProgress: prometheus.NewDesc(
			"miner_seal_sector_task_progress",
			"Miner seal sector task progress",
			[]string{"tasktype", "worker", "sector", "done"}, nil,
		),
		MinerSectorTaskConcurrent: prometheus.NewDesc(
			"miner_seal_sector_task_concurrent_total",
			"Miner seal sector task concurrent total",
			[]string{"tasktype"}, nil,
		),
		MinerSectorTaskDones: prometheus.NewDesc(
			"miner_seal_sector_task_dones_total",
			"Miner seal sector task dones total",
			[]string{"tasktype"}, nil,
		),
	}
	return mm
}

func (m *MinerMetrics) SetHost(host string) {
	m.host = host
}

func (m *MinerMetrics) SetFullnodeHost(host string) {
	m.ml.SetFullnodeHost(host)
}

func (m *MinerMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.ForkBlocks
	ch <- m.PastBlocks
	ch <- m.FailedBlocks
	ch <- m.BlockTookAvgMs
	ch <- m.BlockTookMaxMs
	ch <- m.BlockTookMinMs
	ch <- m.Blocks
	ch <- m.SectorTaskElapsed
	ch <- m.SectorTaskDuration
	ch <- m.SectorTaskConcurrent
	ch <- m.SectorTaskDones
	ch <- m.SectorTaskProgress
	ch <- m.MinerSectorTaskConcurrent
	ch <- m.MinerSectorTaskDones
}

func (m *MinerMetrics) Collect(ch chan<- prometheus.Metric) {
	tooks := m.ml.GetBlockTooks()
	forkBlocks := m.ml.GetForkBlocks()
	pastBlocks := m.ml.GetPastBlocks()
	failedBlocks := m.ml.GetFailedBlocks()

	avgMs := uint64(0)
	maxMs := uint64(0)
	minMs := uint64(0)

	for _, took := range tooks {
		avgMs += took
		if maxMs < took {
			maxMs = took
		}
		if minMs == 0 || took < minMs {
			minMs = took
		}
	}
	if 0 < len(tooks) {
		avgMs = avgMs / uint64(len(tooks))
	}

	ch <- prometheus.MustNewConstMetric(m.ForkBlocks, prometheus.CounterValue, float64(forkBlocks))
	ch <- prometheus.MustNewConstMetric(m.PastBlocks, prometheus.CounterValue, float64(pastBlocks))
	ch <- prometheus.MustNewConstMetric(m.FailedBlocks, prometheus.CounterValue, float64(failedBlocks))
	ch <- prometheus.MustNewConstMetric(m.BlockTookAvgMs, prometheus.CounterValue, float64(avgMs))
	ch <- prometheus.MustNewConstMetric(m.BlockTookMaxMs, prometheus.CounterValue, float64(maxMs))
	ch <- prometheus.MustNewConstMetric(m.BlockTookMinMs, prometheus.CounterValue, float64(minMs))
	ch <- prometheus.MustNewConstMetric(m.Blocks, prometheus.CounterValue, float64(len(tooks)))

	sectorTasks := m.ml.GetSectorTasks()
	totalConcurrent := 0
	totalDones := 0

	for taskType, typedTasks := range sectorTasks {
		for worker, workerTasks := range typedTasks {
			elapsed := uint64(0)
			concurrent := uint64(0)
			duration := uint64(0)
			dones := uint64(0)
			for _, task := range workerTasks {
				if task.Done {
					dones += 1
					totalDones += 1
					if duration < task.Elapsed {
						duration = task.Elapsed
					}
					ch <- prometheus.MustNewConstMetric(m.SectorTaskProgress, prometheus.CounterValue,
						float64(duration), taskType, worker, task.Sector, "1")
				} else {
					concurrent += 1
					totalConcurrent += 1
					if elapsed < task.Elapsed {
						elapsed = task.Elapsed
					}
					ch <- prometheus.MustNewConstMetric(m.SectorTaskProgress, prometheus.CounterValue,
						float64(elapsed), taskType, worker, task.Sector, "0")
				}
			}
			ch <- prometheus.MustNewConstMetric(m.SectorTaskElapsed, prometheus.CounterValue, float64(elapsed), taskType, worker)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskDuration, prometheus.CounterValue, float64(duration), taskType, worker)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskConcurrent, prometheus.CounterValue, float64(concurrent), taskType, worker)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskDones, prometheus.CounterValue, float64(dones), taskType, worker)
		}
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskConcurrent, prometheus.CounterValue, float64(totalConcurrent), taskType)
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskDones, prometheus.CounterValue, float64(totalDones), taskType)
	}
}
