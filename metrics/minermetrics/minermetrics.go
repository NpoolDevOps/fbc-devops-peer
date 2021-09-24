package minermetrics

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/api/minerapi"
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/minerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type MinerMetricsConfig struct {
	ShareStorageRoot string
	Logfile          string
	Username         string
	NetworkType      string
}

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

	Power            *prometheus.Desc
	RawPower         *prometheus.Desc
	CommittedPower   *prometheus.Desc
	ProvingPower     *prometheus.Desc
	FaultyPower      *prometheus.Desc
	MinerBalance     *prometheus.Desc
	PrecommitDeposit *prometheus.Desc
	InitialPledge    *prometheus.Desc
	Vesting          *prometheus.Desc
	Available        *prometheus.Desc
	WorkerBalance    *prometheus.Desc
	ControlBalance   *prometheus.Desc
	MinerTaskState   *prometheus.Desc

	MinerId *prometheus.Desc

	SectorTaskRunning        *prometheus.Desc
	SectorTaskWaiting        *prometheus.Desc
	SectorTaskRunningElapsed *prometheus.Desc
	SectorTaskWaitingElapsed *prometheus.Desc
	MinerSectorTaskRunning   *prometheus.Desc
	MinerSectorTaskWaiting   *prometheus.Desc
	MinerSectorSizeGib       *prometheus.Desc

	MinerBaseFee             *prometheus.Desc
	MinerWorkers             *prometheus.Desc
	MinerGPUs                *prometheus.Desc
	MinerWorkerGPUs          *prometheus.Desc
	MinerWorkerMaintaining   *prometheus.Desc
	MinerWorkerRejectTask    *prometheus.Desc
	MinerCheckSectorsChecked *prometheus.Desc
	MinerCheckSectorsGood    *prometheus.Desc

	ProvingDeadlineAllSectors       *prometheus.Desc
	ProvingDeadlineFaultySectors    *prometheus.Desc
	ProvingDeadlineCurrent          *prometheus.Desc
	ProvingDeadlinePartitions       *prometheus.Desc
	ProvingDeadlineProvenPartitions *prometheus.Desc

	LogFileSize           *prometheus.Desc
	ChainSyncNotCompleted *prometheus.Desc
	ChainNotSuitable      *prometheus.Desc
	ChainHeadListen       *prometheus.Desc

	StorageMountpointPermission *prometheus.Desc
	StorageMountError           *prometheus.Desc

	MinerOpenFileNumber          *prometheus.Desc
	MinerProcessTcpConnectNumber *prometheus.Desc

	MinerAdjustGasFeecap *prometheus.Desc
	MinerAdjustBaseFee   *prometheus.Desc

	MinerIsMaster     *prometheus.Desc
	MinerRepoDirUsage *prometheus.Desc

	MiningLateBase            *prometheus.Desc
	MiningLateWinner          *prometheus.Desc
	MiningLateBaseDeltaSecond *prometheus.Desc
	MiningEligible            *prometheus.Desc
	MiningNetworkPower        *prometheus.Desc
	MiningMinerPower          *prometheus.Desc

	minerInfo   minerapi.MinerInfo
	sealingJobs minerapi.SealingJobs
	workerInfos minerapi.WorkerInfos

	mutex sync.Mutex

	errors           int
	host             string
	hasHost          bool
	fullnodeHost     string
	config           MinerMetricsConfig
	lotusStoragePath []string
	storageStat      map[string]error
	sectorStat       map[string]uint64
	username         string
	networkType      string
}

func NewMinerMetrics(cfg MinerMetricsConfig, paths []string) *MinerMetrics {
	mm := &MinerMetrics{
		ml:               minerlog.NewMinerLog(cfg.Logfile),
		lotusStoragePath: paths,
		config:           cfg,
		username:         cfg.Username,
		networkType:      cfg.NetworkType,
		ForkBlocks: prometheus.NewDesc(
			"miner_fork_blocks",
			"Show miner fork blocks",
			[]string{"networktype", "user"}, nil,
		),
		PastBlocks: prometheus.NewDesc(
			"miner_block_in_past",
			"Show miner block in past",
			[]string{"networktype", "user"}, nil,
		),
		FailedBlocks: prometheus.NewDesc(
			"miner_block_failed",
			"Show miner block failed",
			[]string{"networktype", "user"}, nil,
		),
		BlockTookAvgMs: prometheus.NewDesc(
			"miner_block_took_average_ms",
			"Show miner block took average ms",
			[]string{"networktype", "user"}, nil,
		),
		BlockTookMaxMs: prometheus.NewDesc(
			"miner_block_took_max_ms",
			"Show miner block took max ms",
			[]string{"networktype", "user"}, nil,
		),
		BlockTookMinMs: prometheus.NewDesc(
			"miner_block_took_min_ms",
			"Show miner block took min ms",
			[]string{"networktype", "user"}, nil,
		),
		Blocks: prometheus.NewDesc(
			"miner_block_produced",
			"Show miner block produced",
			[]string{"networktype", "user"}, nil,
		),
		SectorTaskElapsed: prometheus.NewDesc(
			"miner_seal_sector_task_elapsed",
			"Miner seal sector task elapsed",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskDuration: prometheus.NewDesc(
			"miner_seal_sector_task_duration",
			"Miner seal sector task duration",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskConcurrent: prometheus.NewDesc(
			"miner_seal_sector_task_concurrent",
			"Miner seal sector task concurrent",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskDones: prometheus.NewDesc(
			"miner_seal_sector_task_dones",
			"Miner seal sector task dones",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskProgress: prometheus.NewDesc(
			"miner_seal_sector_task_progress",
			"Miner seal sector task progress",
			[]string{"tasktype", "worker", "sector", "done", "networktype", "user"}, nil,
		),
		MinerSectorTaskConcurrent: prometheus.NewDesc(
			"miner_seal_sector_task_concurrent_total",
			"Miner seal sector task concurrent total",
			[]string{"tasktype", "networktype", "user"}, nil,
		),
		MinerSectorTaskDones: prometheus.NewDesc(
			"miner_seal_sector_task_dones_total",
			"Miner seal sector task dones total",
			[]string{"tasktype", "networktype", "user"}, nil,
		),
		Power: prometheus.NewDesc(
			"miner_power",
			"Miner power",
			[]string{"networktype", "user"}, nil,
		),
		RawPower: prometheus.NewDesc(
			"miner_raw_power",
			"Miner raw power",
			[]string{"networktype", "user"}, nil,
		),
		CommittedPower: prometheus.NewDesc(
			"miner_committed_power",
			"Miner committed power",
			[]string{"networktype", "user"}, nil,
		),
		ProvingPower: prometheus.NewDesc(
			"miner_proving_power",
			"Miner proving power",
			[]string{"networktype", "user"}, nil,
		),
		FaultyPower: prometheus.NewDesc(
			"miner_faulty_power",
			"Miner faulty power",
			[]string{"networktype", "user"}, nil,
		),
		MinerBalance: prometheus.NewDesc(
			"miner_balance",
			"Miner balance",
			[]string{"networktype", "user"}, nil,
		),
		PrecommitDeposit: prometheus.NewDesc(
			"miner_precommit_deposit",
			"Miner precommit deposit",
			[]string{"networktype", "user"}, nil,
		),
		InitialPledge: prometheus.NewDesc(
			"miner_initial_pledge",
			"Miner initial pledge",
			[]string{"networktype", "user"}, nil,
		),
		Vesting: prometheus.NewDesc(
			"miner_vesting",
			"Miner vesting",
			[]string{"networktype", "user"}, nil,
		),
		Available: prometheus.NewDesc(
			"miner_available",
			"Miner available",
			[]string{"networktype", "user"}, nil,
		),
		WorkerBalance: prometheus.NewDesc(
			"miner_worker_balance",
			"Miner worker balance",
			[]string{"networktype", "user"}, nil,
		),
		ControlBalance: prometheus.NewDesc(
			"miner_control_balance",
			"Miner control balance",
			[]string{"networktype", "user"}, nil,
		),
		MinerTaskState: prometheus.NewDesc(
			"miner_sector_state",
			"Miner sector state",
			[]string{"state", "networktype", "user"}, nil,
		),
		SectorTaskRunning: prometheus.NewDesc(
			"miner_sector_task_running",
			"Miner sector task running",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskRunningElapsed: prometheus.NewDesc(
			"miner_sector_task_running_elapsed",
			"Miner sector task running elapsed",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskWaiting: prometheus.NewDesc(
			"miner_sector_task_waiting",
			"Miner sector task waiting",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		SectorTaskWaitingElapsed: prometheus.NewDesc(
			"miner_sector_task_waiting_elapsed",
			"Miner sector task waiting elapsed",
			[]string{"tasktype", "worker", "networktype", "user"}, nil,
		),
		MinerSectorTaskRunning: prometheus.NewDesc(
			"miner_sector_task_running_total",
			"Miner sector task running total",
			[]string{"tasktype", "networktype", "user"}, nil,
		),
		MinerSectorTaskWaiting: prometheus.NewDesc(
			"miner_sector_task_waiting_total",
			"Miner sector task waiting total",
			[]string{"tasktype", "networktype", "user"}, nil,
		),
		MinerBaseFee: prometheus.NewDesc(
			"miner_basefee",
			"Miner basefee",
			[]string{"networktype", "user"}, nil,
		),
		MinerWorkers: prometheus.NewDesc(
			"miner_workers",
			"Miner workers",
			[]string{"networktype", "user"}, nil,
		),
		MinerWorkerGPUs: prometheus.NewDesc(
			"miner_worker_gpus",
			"Miner worker gpus",
			[]string{"worker", "networktype", "user"}, nil,
		),
		MinerGPUs: prometheus.NewDesc(
			"miner_gpus",
			"Miner gpus",
			[]string{"networktype", "user"}, nil,
		),
		MinerWorkerMaintaining: prometheus.NewDesc(
			"miner_worker_maintaining",
			"Miner worker maintaining",
			[]string{"worker", "networktype", "user"}, nil,
		),
		MinerWorkerRejectTask: prometheus.NewDesc(
			"miner_worker_reject_task",
			"Miner worker reject task",
			[]string{"worker", "networktype", "user"}, nil,
		),
		MinerCheckSectorsGood: prometheus.NewDesc(
			"miner_check_sectors_good",
			"Miner check sectors good",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		MinerCheckSectorsChecked: prometheus.NewDesc(
			"miner_check_sectors_checked",
			"Miner check sectors checked",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		ProvingDeadlineAllSectors: prometheus.NewDesc(
			"miner_proving_deadline_all_sectors",
			"Miner proving deadline all sectors",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		ProvingDeadlineFaultySectors: prometheus.NewDesc(
			"miner_proving_deadline_faulty_sectors",
			"Miner proving deadline faulty sectors",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		ProvingDeadlineCurrent: prometheus.NewDesc(
			"miner_proving_deadline_current",
			"Miner proving deadline current",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		ProvingDeadlinePartitions: prometheus.NewDesc(
			"miner_proving_deadline_partitions",
			"Miner proving deadline partitions",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		ProvingDeadlineProvenPartitions: prometheus.NewDesc(
			"miner_proving_deadline_proven_partitions",
			"Miner proving deadline proven partitions",
			[]string{"deadline", "networktype", "user"}, nil,
		),
		LogFileSize: prometheus.NewDesc(
			"miner_log_filesize",
			"Miner log filesize",
			[]string{"networktype", "user"}, nil,
		),
		ChainSyncNotCompleted: prometheus.NewDesc(
			"miner_chain_sync_not_completed",
			"Miner chain sync not completed",
			[]string{"fullnode", "networktype", "user"}, nil,
		),
		ChainNotSuitable: prometheus.NewDesc(
			"miner_chain_not_suitable",
			"Miner chain not suitable",
			[]string{"networktype", "user"}, nil,
		),
		ChainHeadListen: prometheus.NewDesc(
			"miner_chain_head_epoch",
			"Miner chain head epoch",
			[]string{"fullnode", "networktype", "user"}, nil,
		),
		StorageMountpointPermission: prometheus.NewDesc(
			"miner_storage_mount_point_permission",
			"show miner storage's file mount point permission",
			[]string{"filedir", "networktype", "user"}, nil,
		),
		StorageMountError: prometheus.NewDesc(
			"miner_storage_mount_error",
			"show storage mount error",
			[]string{"filedir", "networktype", "user"}, nil,
		),
		MinerOpenFileNumber: prometheus.NewDesc(
			"miner_open_file_number",
			"show how many files miner opened",
			[]string{"networktype", "user"}, nil,
		),
		MinerProcessTcpConnectNumber: prometheus.NewDesc(
			"miner_process_tcp_connect_number",
			"show miner process tcp connect number",
			[]string{"networktype", "user"}, nil,
		),
		MinerAdjustGasFeecap: prometheus.NewDesc(
			"miner_fee_adjust_gas_feecap",
			"show miner fee adjust gas feecap",
			[]string{"networktype", "user"}, nil,
		),
		MinerAdjustBaseFee: prometheus.NewDesc(
			"miner_fee_adjust_basefee",
			"show miner fee adjust base fee",
			[]string{"networktype", "user"}, nil,
		),
		MinerSectorSizeGib: prometheus.NewDesc(
			"miner_sector_size_gib",
			"show miner sector size Gib",
			[]string{"networktype", "user"}, nil,
		),
		MinerIsMaster: prometheus.NewDesc(
			"miner_is_master",
			"show whether miner is master",
			[]string{"networktype", "user"}, nil,
		),
		MiningLateBase: prometheus.NewDesc(
			"miner_mining_late_base",
			"show mining late base",
			[]string{"networktype", "user"}, nil,
		),
		MiningLateBaseDeltaSecond: prometheus.NewDesc(
			"miner_mining_late_base_delta_second",
			"show mining late base delta second",
			[]string{"networktype", "user"}, nil,
		),
		MiningLateWinner: prometheus.NewDesc(
			"miner_mining_late_winner",
			"show mining late winner",
			[]string{"networktype", "user"}, nil,
		),
		MiningEligible: prometheus.NewDesc(
			"miner_mining_eligible",
			"show mining eligible",
			[]string{"networktype", "user"}, nil,
		),
		MiningNetworkPower: prometheus.NewDesc(
			"miner_mining_network_power",
			"show mining network power",
			[]string{"networktype", "user"}, nil,
		),
		MiningMinerPower: prometheus.NewDesc(
			"miner_mining_miner_power",
			"show mining miner power",
			[]string{"networktype", "user"}, nil,
		),
		MinerId: prometheus.NewDesc(
			"miner_id",
			"show miner id",
			[]string{"minerid", "networktype", "user"}, nil,
		),
		MinerRepoDirUsage: prometheus.NewDesc(
			"miner_repo_dir_usage",
			"show miner repo dir usage",
			[]string{"repodir", "totalcap", "networktype", "user"}, nil,
		),
	}

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		infoCh := make(chan minerapi.MinerInfo)
		jobsCh := make(chan minerapi.SealingJobs)
		workersCh := make(chan minerapi.WorkerInfos)
		count := 0
		for {
			showSectors := false
			if count%15 == 0 {
				showSectors = true
			}

			count += 1

			minerapi.GetMinerInfo(infoCh, showSectors)
			info := <-infoCh

			if len(info.State) != 0 {
				mm.mutex.Lock()
				mm.sectorStat = info.State
				mm.mutex.Unlock()
			}

			mm.mutex.Lock()
			mm.minerInfo = info
			mm.mutex.Unlock()

			minerapi.GetSealingJobs(jobsCh)
			jobs := <-jobsCh

			mm.mutex.Lock()
			mm.sealingJobs = jobs
			mm.mutex.Unlock()

			minerapi.GetWorkerInfos(workersCh)
			workerInfos := <-workersCh

			mm.mutex.Lock()
			mm.workerInfos = workerInfos
			mm.mutex.Unlock()

			storageStat := systemapi.StatSubDirs(cfg.ShareStorageRoot, 1)
			mm.mutex.Lock()
			mm.storageStat = storageStat
			mm.mutex.Unlock()

			<-ticker.C
		}
	}()

	return mm
}

func (m *MinerMetrics) SetHost(host string) {
	m.host = host
	m.hasHost = true
}

func (m *MinerMetrics) SetFullnodeHost(host string) {
	m.fullnodeHost = host
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
	ch <- m.Power
	ch <- m.RawPower
	ch <- m.CommittedPower
	ch <- m.ProvingPower
	ch <- m.FaultyPower
	ch <- m.MinerBalance
	ch <- m.PrecommitDeposit
	ch <- m.InitialPledge
	ch <- m.Vesting
	ch <- m.Available
	ch <- m.WorkerBalance
	ch <- m.ControlBalance
	ch <- m.MinerTaskState
	ch <- m.SectorTaskRunning
	ch <- m.SectorTaskWaiting
	ch <- m.SectorTaskRunningElapsed
	ch <- m.SectorTaskWaitingElapsed
	ch <- m.MinerSectorTaskRunning
	ch <- m.MinerSectorTaskWaiting
	ch <- m.MinerBaseFee
	ch <- m.MinerWorkers
	ch <- m.MinerGPUs
	ch <- m.MinerWorkerGPUs
	ch <- m.MinerWorkerMaintaining
	ch <- m.MinerWorkerRejectTask
	ch <- m.MinerCheckSectorsGood
	ch <- m.MinerCheckSectorsChecked
	ch <- m.ProvingDeadlineAllSectors
	ch <- m.ProvingDeadlineFaultySectors
	ch <- m.ProvingDeadlineCurrent
	ch <- m.ProvingDeadlinePartitions
	ch <- m.ProvingDeadlineProvenPartitions
	ch <- m.LogFileSize
	ch <- m.ChainSyncNotCompleted
	ch <- m.ChainNotSuitable
	ch <- m.ChainHeadListen
	ch <- m.StorageMountpointPermission
	ch <- m.StorageMountError
	ch <- m.MinerOpenFileNumber
	ch <- m.MinerProcessTcpConnectNumber
	ch <- m.MinerAdjustBaseFee
	ch <- m.MinerAdjustGasFeecap
	ch <- m.MinerSectorSizeGib
	ch <- m.MinerIsMaster
	ch <- m.MiningEligible
	ch <- m.MiningLateBase
	ch <- m.MiningLateBaseDeltaSecond
	ch <- m.MiningLateWinner
	ch <- m.MiningMinerPower
	ch <- m.MiningNetworkPower
	ch <- m.MinerId

}

func (m *MinerMetrics) Collect(ch chan<- prometheus.Metric) {
	tooks := m.ml.GetBlockTooks()
	forkBlocks := m.ml.GetForkBlocks()
	pastBlocks := m.ml.GetPastBlocks()
	failedBlocks := m.ml.GetFailedBlocks()
	minerAdjustGasFeecap := m.ml.GetMinerFeeAdjustGasFeecap()
	minerAdjustBaseFee := m.ml.GetMinerAdjustBaseFee()
	minerIsMaster := m.ml.GetMinerIsMaster()
	mineOne := m.ml.GetMineOne()
	username := m.username
	networkType := m.networkType

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

	ch <- prometheus.MustNewConstMetric(m.ForkBlocks, prometheus.CounterValue, float64(forkBlocks), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.PastBlocks, prometheus.CounterValue, float64(pastBlocks), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.FailedBlocks, prometheus.CounterValue, float64(failedBlocks), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.BlockTookAvgMs, prometheus.CounterValue, float64(avgMs), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.BlockTookMaxMs, prometheus.CounterValue, float64(maxMs), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.BlockTookMinMs, prometheus.CounterValue, float64(minMs), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.Blocks, prometheus.CounterValue, float64(len(tooks)), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.MinerIsMaster, prometheus.CounterValue, minerIsMaster, networkType, username)

	sectorTasks := m.ml.GetSectorTasks()
	for taskType, typedTasks := range sectorTasks {
		totalConcurrent := uint64(0)
		totalDones := uint64(0)
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
						float64(task.Elapsed), taskType, worker, task.Sector, "1", networkType, username)
				} else {
					concurrent += 1
					totalConcurrent += 1
					if elapsed < task.Elapsed && task.Elapsed < uint64(24*time.Hour.Seconds()) {
						elapsed = task.Elapsed
					}
					ch <- prometheus.MustNewConstMetric(m.SectorTaskProgress, prometheus.CounterValue,
						float64(task.Elapsed), taskType, worker, task.Sector, "0", networkType, username)
				}
			}
			ch <- prometheus.MustNewConstMetric(m.SectorTaskElapsed, prometheus.CounterValue, float64(elapsed), taskType, worker, networkType, username)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskDuration, prometheus.CounterValue, float64(duration), taskType, worker, networkType, username)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskConcurrent, prometheus.CounterValue, float64(concurrent), taskType, worker, networkType, username)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskDones, prometheus.CounterValue, float64(dones), taskType, worker, networkType, username)
		}
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskConcurrent, prometheus.CounterValue, float64(totalConcurrent), taskType, networkType, username)
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskDones, prometheus.CounterValue, float64(totalDones), taskType, networkType, username)
	}

	m.mutex.Lock()
	jobs := m.sealingJobs
	m.mutex.Unlock()
	for taskType, typedJobs := range jobs.Jobs {
		totalRunning := uint64(0)
		totalWaiting := uint64(0)
		for worker, job := range typedJobs {
			ch <- prometheus.MustNewConstMetric(m.SectorTaskWaitingElapsed, prometheus.CounterValue, float64(job.MaxWaiting), taskType, worker, networkType, username)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskRunningElapsed, prometheus.CounterValue, float64(job.MaxRunning), taskType, worker, networkType, username)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskRunning, prometheus.CounterValue, float64(job.Running), taskType, worker, networkType, username)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskWaiting, prometheus.CounterValue, float64(job.Assigned), taskType, worker, networkType, username)
			totalRunning += job.Running
			totalWaiting += job.Assigned
		}
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskRunning, prometheus.CounterValue, float64(totalRunning), taskType, networkType, username)
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskWaiting, prometheus.CounterValue, float64(totalWaiting), taskType, networkType, username)
	}

	m.mutex.Lock()
	info := m.minerInfo
	m.mutex.Unlock()
	ch <- prometheus.MustNewConstMetric(m.Power, prometheus.CounterValue, float64(info.Power), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.RawPower, prometheus.CounterValue, float64(info.Raw), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.CommittedPower, prometheus.CounterValue, float64(info.Committed), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.ProvingPower, prometheus.CounterValue, float64(info.Proving), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.FaultyPower, prometheus.CounterValue, float64(info.Faulty), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.MinerBalance, prometheus.CounterValue, float64(info.MinerBalance), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.PrecommitDeposit, prometheus.CounterValue, float64(info.PrecommitDeposit), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.InitialPledge, prometheus.CounterValue, float64(info.InitialPledge), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.Vesting, prometheus.CounterValue, float64(info.Vesting), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.Available, prometheus.CounterValue, float64(info.Available), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.WorkerBalance, prometheus.CounterValue, float64(info.WorkerBalance), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.ControlBalance, prometheus.CounterValue, float64(info.ControlBalance), networkType, username)
	for state, count := range m.sectorStat {
		ch <- prometheus.MustNewConstMetric(m.MinerTaskState, prometheus.CounterValue, float64(count), state, networkType, username)
	}
	ch <- prometheus.MustNewConstMetric(m.MinerSectorSizeGib, prometheus.CounterValue, float64(info.SectorSize), networkType, username)

	basefee, _ := lotusapi.ChainBaseFee(m.fullnodeHost)
	ch <- prometheus.MustNewConstMetric(m.MinerBaseFee, prometheus.CounterValue, basefee, networkType, username)

	m.mutex.Lock()
	workerInfos := m.workerInfos
	m.mutex.Unlock()

	gpus := 0
	ch <- prometheus.MustNewConstMetric(m.MinerWorkers, prometheus.CounterValue, float64(len(workerInfos.Infos)), networkType, username)
	for worker, info := range workerInfos.Infos {
		ch <- prometheus.MustNewConstMetric(m.MinerWorkerGPUs, prometheus.CounterValue, float64(info.GPUs), worker, networkType, username)
		ch <- prometheus.MustNewConstMetric(m.MinerWorkerMaintaining, prometheus.CounterValue, float64(info.Maintaining), worker, networkType, username)
		ch <- prometheus.MustNewConstMetric(m.MinerWorkerRejectTask, prometheus.CounterValue, float64(info.RejectTask), worker, networkType, username)
		gpus += info.GPUs
	}
	ch <- prometheus.MustNewConstMetric(m.MinerGPUs, prometheus.CounterValue, float64(gpus), networkType, username)

	checkSectors := m.ml.GetCheckSectors()
	for deadline, sectors := range checkSectors {
		ch <- prometheus.MustNewConstMetric(m.MinerCheckSectorsGood, prometheus.CounterValue, float64(sectors.Good), fmt.Sprintf("%v", deadline), networkType, username)
		ch <- prometheus.MustNewConstMetric(m.MinerCheckSectorsChecked, prometheus.CounterValue, float64(sectors.Checked), fmt.Sprintf("%v", deadline), networkType, username)
	}
	m.mutex.Lock()
	minerId := m.minerInfo.MinerId
	m.mutex.Unlock()

	if 0 < len(minerId) {
		ch <- prometheus.MustNewConstMetric(m.MinerId, prometheus.CounterValue, float64(1), minerId, networkType, username)
		deadlines, err := lotusapi.ProvingDeadlines(m.fullnodeHost, minerId)
		if err == nil {
			for dlIdx, deadline := range deadlines.Deadlines {
				current := 0
				if deadline.Current {
					current = 1
				}
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineAllSectors, prometheus.CounterValue, float64(deadline.AllSectors), fmt.Sprintf("%v", dlIdx), networkType, username)
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineFaultySectors, prometheus.CounterValue, float64(deadline.FaultySectors), fmt.Sprintf("%v", dlIdx), networkType, username)
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineCurrent, prometheus.CounterValue, float64(current), fmt.Sprintf("%v", dlIdx), networkType, username)
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlinePartitions, prometheus.CounterValue, float64(deadline.Partitions), fmt.Sprintf("%v", dlIdx), networkType, username)
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineProvenPartitions, prometheus.CounterValue, float64(deadline.ProvenPartitions), fmt.Sprintf("%v", dlIdx), networkType, username)
			}
		} else {
			log.Errorf(log.Fields{}, "fail to get proving deadlines: %v", err)
		}
	}

	filesize := m.ml.LogFileSize()
	ch <- prometheus.MustNewConstMetric(m.LogFileSize, prometheus.CounterValue, float64(filesize), networkType, username)

	chainSyncNotCompletedHosts := m.ml.GetChainSyncNotCompletedHosts()
	//
	for host := range chainSyncNotCompletedHosts {
		ch <- prometheus.MustNewConstMetric(m.ChainSyncNotCompleted, prometheus.CounterValue, float64(1), host, networkType, username)
	}
	chainNotSuitable := m.ml.GetChainNotSuitable()
	ch <- prometheus.MustNewConstMetric(m.ChainNotSuitable, prometheus.CounterValue, float64(chainNotSuitable), networkType, username)

	chainHeadListenSuccessHosts := m.ml.GetChainHeadListenSuccessHosts()
	for host, epoch := range chainHeadListenSuccessHosts {
		ch <- prometheus.MustNewConstMetric(m.ChainHeadListen, prometheus.CounterValue, float64(epoch), host, networkType, username)
	}

	for k, v := range m.storageStat {
		if v != nil {
			ch <- prometheus.MustNewConstMetric(m.StorageMountError, prometheus.CounterValue, 1, k, networkType, username)
		} else {
			ch <- prometheus.MustNewConstMetric(m.StorageMountError, prometheus.CounterValue, 0, k, networkType, username)
		}
		filePerm, _ := systemapi.FilePerm2Int(k)
		ch <- prometheus.MustNewConstMetric(m.StorageMountpointPermission, prometheus.CounterValue, float64(filePerm), k, networkType, username)
	}

	minerFileOpenNumber, _ := systemapi.GetProcessOpenFileNumber("lotus-miner")
	ch <- prometheus.MustNewConstMetric(m.MinerOpenFileNumber, prometheus.CounterValue, float64(minerFileOpenNumber), networkType, username)

	tcpConnectNumber, _ := systemapi.GetProcessTcpConnectNumber("lotus-miner")
	ch <- prometheus.MustNewConstMetric(m.MinerProcessTcpConnectNumber, prometheus.CounterValue, float64(tcpConnectNumber), networkType, username)

	ch <- prometheus.MustNewConstMetric(m.MinerAdjustBaseFee, prometheus.CounterValue, minerAdjustBaseFee, networkType, username)
	ch <- prometheus.MustNewConstMetric(m.MinerAdjustGasFeecap, prometheus.CounterValue, minerAdjustGasFeecap, networkType, username)

	if mineOne.MiningEligible {
		ch <- prometheus.MustNewConstMetric(m.MiningEligible, prometheus.CounterValue, float64(1), networkType, username)
	} else {
		ch <- prometheus.MustNewConstMetric(m.MiningEligible, prometheus.CounterValue, float64(0), networkType, username)
	}
	if mineOne.MiningLateBase {
		ch <- prometheus.MustNewConstMetric(m.MiningLateBase, prometheus.CounterValue, float64(1), networkType, username)
	} else {
		ch <- prometheus.MustNewConstMetric(m.MiningLateBase, prometheus.CounterValue, float64(0), networkType, username)
	}
	if mineOne.MiningLateWinner {
		ch <- prometheus.MustNewConstMetric(m.MiningLateWinner, prometheus.CounterValue, float64(1), networkType, username)
	} else {
		ch <- prometheus.MustNewConstMetric(m.MiningLateWinner, prometheus.CounterValue, float64(0), networkType, username)
	}
	switch mineOne.MiningLateBaseDeltaSecond.(type) {
	case float64:
		ch <- prometheus.MustNewConstMetric(m.MiningLateBaseDeltaSecond, prometheus.CounterValue, mineOne.MiningLateBaseDeltaSecond.(float64), networkType, username)
	case int64:
		ch <- prometheus.MustNewConstMetric(m.MiningLateBaseDeltaSecond, prometheus.CounterValue, float64(mineOne.MiningLateBaseDeltaSecond.(int64)), networkType, username)
	case string:
		miningLateBaseDeltaSecond, _ := strconv.ParseFloat(mineOne.MiningLateBaseDeltaSecond.(string), 64)
		ch <- prometheus.MustNewConstMetric(m.MiningLateBaseDeltaSecond, prometheus.CounterValue, miningLateBaseDeltaSecond, networkType, username)
	}
	miningNetworkPower, _ := strconv.ParseFloat(mineOne.MiningNetworkPower, 64)
	miningMinerPower, _ := strconv.ParseFloat(mineOne.MiningMinerPower, 64)
	ch <- prometheus.MustNewConstMetric(m.MiningNetworkPower, prometheus.CounterValue, miningNetworkPower, networkType, username)
	ch <- prometheus.MustNewConstMetric(m.MiningMinerPower, prometheus.CounterValue, miningMinerPower, networkType, username)

	for _, path := range m.lotusStoragePath {
		pathStatus := getMinerRepoDirUsage(path)
		ch <- prometheus.MustNewConstMetric(m.MinerRepoDirUsage, prometheus.CounterValue, pathStatus.Used, fmt.Sprintf("%v", path), fmt.Sprintf("%v", pathStatus.All), networkType, username)
	}
}

func getMinerRepoDirUsage(dir string) systemapi.DiskStatus {
	return systemapi.DiskUsage(dir)
}
