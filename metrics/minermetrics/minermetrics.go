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
}

func NewMinerMetrics(cfg MinerMetricsConfig, paths []string) *MinerMetrics {
	mm := &MinerMetrics{
		ml:               minerlog.NewMinerLog(cfg.Logfile),
		lotusStoragePath: paths,
		config:           cfg,
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
		Power: prometheus.NewDesc(
			"miner_power",
			"Miner power",
			nil, nil,
		),
		RawPower: prometheus.NewDesc(
			"miner_raw_power",
			"Miner raw power",
			nil, nil,
		),
		CommittedPower: prometheus.NewDesc(
			"miner_committed_power",
			"Miner committed power",
			nil, nil,
		),
		ProvingPower: prometheus.NewDesc(
			"miner_proving_power",
			"Miner proving power",
			nil, nil,
		),
		FaultyPower: prometheus.NewDesc(
			"miner_faulty_power",
			"Miner faulty power",
			nil, nil,
		),
		MinerBalance: prometheus.NewDesc(
			"miner_balance",
			"Miner balance",
			nil, nil,
		),
		PrecommitDeposit: prometheus.NewDesc(
			"miner_precommit_deposit",
			"Miner precommit deposit",
			nil, nil,
		),
		InitialPledge: prometheus.NewDesc(
			"miner_initial_pledge",
			"Miner initial pledge",
			nil, nil,
		),
		Vesting: prometheus.NewDesc(
			"miner_vesting",
			"Miner vesting",
			nil, nil,
		),
		Available: prometheus.NewDesc(
			"miner_available",
			"Miner available",
			nil, nil,
		),
		WorkerBalance: prometheus.NewDesc(
			"miner_worker_balance",
			"Miner worker balance",
			nil, nil,
		),
		ControlBalance: prometheus.NewDesc(
			"miner_control_balance",
			"Miner control balance",
			nil, nil,
		),
		MinerTaskState: prometheus.NewDesc(
			"miner_sector_state",
			"Miner sector state",
			[]string{"state"}, nil,
		),
		SectorTaskRunning: prometheus.NewDesc(
			"miner_sector_task_running",
			"Miner sector task running",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskRunningElapsed: prometheus.NewDesc(
			"miner_sector_task_running_elapsed",
			"Miner sector task running elapsed",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskWaiting: prometheus.NewDesc(
			"miner_sector_task_waiting",
			"Miner sector task waiting",
			[]string{"tasktype", "worker"}, nil,
		),
		SectorTaskWaitingElapsed: prometheus.NewDesc(
			"miner_sector_task_waiting_elapsed",
			"Miner sector task waiting elapsed",
			[]string{"tasktype", "worker"}, nil,
		),
		MinerSectorTaskRunning: prometheus.NewDesc(
			"miner_sector_task_running_total",
			"Miner sector task running total",
			[]string{"tasktype"}, nil,
		),
		MinerSectorTaskWaiting: prometheus.NewDesc(
			"miner_sector_task_waiting_total",
			"Miner sector task waiting total",
			[]string{"tasktype"}, nil,
		),
		MinerBaseFee: prometheus.NewDesc(
			"miner_basefee",
			"Miner basefee",
			nil, nil,
		),
		MinerWorkers: prometheus.NewDesc(
			"miner_workers",
			"Miner workers",
			nil, nil,
		),
		MinerWorkerGPUs: prometheus.NewDesc(
			"miner_worker_gpus",
			"Miner worker gpus",
			[]string{"worker"}, nil,
		),
		MinerGPUs: prometheus.NewDesc(
			"miner_gpus",
			"Miner gpus",
			nil, nil,
		),
		MinerWorkerMaintaining: prometheus.NewDesc(
			"miner_worker_maintaining",
			"Miner worker maintaining",
			[]string{"worker"}, nil,
		),
		MinerWorkerRejectTask: prometheus.NewDesc(
			"miner_worker_reject_task",
			"Miner worker reject task",
			[]string{"worker"}, nil,
		),
		MinerCheckSectorsGood: prometheus.NewDesc(
			"miner_check_sectors_good",
			"Miner check sectors good",
			[]string{"deadline"}, nil,
		),
		MinerCheckSectorsChecked: prometheus.NewDesc(
			"miner_check_sectors_checked",
			"Miner check sectors checked",
			[]string{"deadline"}, nil,
		),
		ProvingDeadlineAllSectors: prometheus.NewDesc(
			"miner_proving_deadline_all_sectors",
			"Miner proving deadline all sectors",
			[]string{"deadline"}, nil,
		),
		ProvingDeadlineFaultySectors: prometheus.NewDesc(
			"miner_proving_deadline_faulty_sectors",
			"Miner proving deadline faulty sectors",
			[]string{"deadline"}, nil,
		),
		ProvingDeadlineCurrent: prometheus.NewDesc(
			"miner_proving_deadline_current",
			"Miner proving deadline current",
			[]string{"deadline"}, nil,
		),
		ProvingDeadlinePartitions: prometheus.NewDesc(
			"miner_proving_deadline_partitions",
			"Miner proving deadline partitions",
			[]string{"deadline"}, nil,
		),
		ProvingDeadlineProvenPartitions: prometheus.NewDesc(
			"miner_proving_deadline_proven_partitions",
			"Miner proving deadline proven partitions",
			[]string{"deadline"}, nil,
		),
		LogFileSize: prometheus.NewDesc(
			"miner_log_filesize",
			"Miner log filesize",
			nil, nil,
		),
		ChainSyncNotCompleted: prometheus.NewDesc(
			"miner_chain_sync_not_completed",
			"Miner chain sync not completed",
			[]string{"fullnode"}, nil,
		),
		ChainNotSuitable: prometheus.NewDesc(
			"miner_chain_not_suitable",
			"Miner chain not suitable",
			nil, nil,
		),
		ChainHeadListen: prometheus.NewDesc(
			"miner_chain_head_epoch",
			"Miner chain head epoch",
			[]string{"fullnode"}, nil,
		),
		StorageMountpointPermission: prometheus.NewDesc(
			"miner_storage_mount_point_permission",
			"show miner storage's file mount point permission",
			[]string{"filedir"}, nil,
		),
		StorageMountError: prometheus.NewDesc(
			"miner_storage_mount_error",
			"show storage mount error",
			[]string{"filedir"}, nil,
		),
		MinerOpenFileNumber: prometheus.NewDesc(
			"miner_open_file_number",
			"show how many files miner opened",
			nil, nil,
		),
		MinerProcessTcpConnectNumber: prometheus.NewDesc(
			"miner_process_tcp_connect_number",
			"show miner process tcp connect number",
			nil, nil,
		),
		MinerAdjustGasFeecap: prometheus.NewDesc(
			"miner_fee_adjust_gas_feecap",
			"show miner fee adjust gas feecap",
			nil, nil,
		),
		MinerAdjustBaseFee: prometheus.NewDesc(
			"miner_fee_adjust_basefee",
			"show miner fee adjust base fee",
			nil, nil,
		),
		MinerSectorSizeGib: prometheus.NewDesc(
			"miner_sector_size_gib",
			"show miner sector size Gib",
			nil, nil,
		),
		MinerIsMaster: prometheus.NewDesc(
			"miner_is_master",
			"show whether miner is master",
			nil, nil,
		),
		MiningLateBase: prometheus.NewDesc(
			"miner_mining_late_base",
			"show mining late base",
			nil, nil,
		),
		MiningLateBaseDeltaSecond: prometheus.NewDesc(
			"miner_mining_late_base_delta_second",
			"show mining late base delta second",
			nil, nil,
		),
		MiningLateWinner: prometheus.NewDesc(
			"miner_mining_late_winner",
			"show mining late winner",
			nil, nil,
		),
		MiningEligible: prometheus.NewDesc(
			"miner_mining_eligible",
			"show mining eligible",
			nil, nil,
		),
		MiningNetworkPower: prometheus.NewDesc(
			"miner_mining_network_power",
			"show mining network power",
			nil, nil,
		),
		MiningMinerPower: prometheus.NewDesc(
			"miner_mining_miner_power",
			"show mining miner power",
			nil, nil,
		),
		MinerId: prometheus.NewDesc(
			"miner_id",
			"show miner id",
			[]string{"minerid"}, nil,
		),
		MinerRepoDirUsage: prometheus.NewDesc(
			"miner_repo_dir_usage",
			"show miner repo dir usage",
			[]string{"repodir", "totalcap"}, nil,
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
	ch <- prometheus.MustNewConstMetric(m.MinerIsMaster, prometheus.CounterValue, minerIsMaster)

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
						float64(task.Elapsed), taskType, worker, task.Sector, "1")
				} else {
					concurrent += 1
					totalConcurrent += 1
					if elapsed < task.Elapsed {
						elapsed = task.Elapsed
					}
					ch <- prometheus.MustNewConstMetric(m.SectorTaskProgress, prometheus.CounterValue,
						float64(task.Elapsed), taskType, worker, task.Sector, "0")
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

	m.mutex.Lock()
	jobs := m.sealingJobs
	m.mutex.Unlock()
	for taskType, typedJobs := range jobs.Jobs {
		totalRunning := uint64(0)
		totalWaiting := uint64(0)
		for worker, job := range typedJobs {
			ch <- prometheus.MustNewConstMetric(m.SectorTaskWaitingElapsed, prometheus.CounterValue, float64(job.MaxWaiting), taskType, worker)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskRunningElapsed, prometheus.CounterValue, float64(job.MaxRunning), taskType, worker)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskRunning, prometheus.CounterValue, float64(job.Running), taskType, worker)
			ch <- prometheus.MustNewConstMetric(m.SectorTaskWaiting, prometheus.CounterValue, float64(job.Assigned), taskType, worker)
			totalRunning += job.Running
			totalWaiting += job.Assigned
		}
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskRunning, prometheus.CounterValue, float64(totalRunning), taskType)
		ch <- prometheus.MustNewConstMetric(m.MinerSectorTaskWaiting, prometheus.CounterValue, float64(totalWaiting), taskType)
	}

	m.mutex.Lock()
	info := m.minerInfo
	m.mutex.Unlock()
	ch <- prometheus.MustNewConstMetric(m.Power, prometheus.CounterValue, float64(info.Power))
	ch <- prometheus.MustNewConstMetric(m.RawPower, prometheus.CounterValue, float64(info.Raw))
	ch <- prometheus.MustNewConstMetric(m.CommittedPower, prometheus.CounterValue, float64(info.Committed))
	ch <- prometheus.MustNewConstMetric(m.ProvingPower, prometheus.CounterValue, float64(info.Proving))
	ch <- prometheus.MustNewConstMetric(m.FaultyPower, prometheus.CounterValue, float64(info.Faulty))
	ch <- prometheus.MustNewConstMetric(m.MinerBalance, prometheus.CounterValue, float64(info.MinerBalance))
	ch <- prometheus.MustNewConstMetric(m.PrecommitDeposit, prometheus.CounterValue, float64(info.PrecommitDeposit))
	ch <- prometheus.MustNewConstMetric(m.InitialPledge, prometheus.CounterValue, float64(info.InitialPledge))
	ch <- prometheus.MustNewConstMetric(m.Vesting, prometheus.CounterValue, float64(info.Vesting))
	ch <- prometheus.MustNewConstMetric(m.Available, prometheus.CounterValue, float64(info.Available))
	ch <- prometheus.MustNewConstMetric(m.WorkerBalance, prometheus.CounterValue, float64(info.WorkerBalance))
	ch <- prometheus.MustNewConstMetric(m.ControlBalance, prometheus.CounterValue, float64(info.ControlBalance))
	for state, count := range m.sectorStat {
		ch <- prometheus.MustNewConstMetric(m.MinerTaskState, prometheus.CounterValue, float64(count), state)
	}
	ch <- prometheus.MustNewConstMetric(m.MinerSectorSizeGib, prometheus.CounterValue, float64(info.SectorSize))

	basefee, _ := lotusapi.ChainBaseFee(m.fullnodeHost)
	ch <- prometheus.MustNewConstMetric(m.MinerBaseFee, prometheus.CounterValue, basefee)

	m.mutex.Lock()
	workerInfos := m.workerInfos
	m.mutex.Unlock()

	gpus := 0
	ch <- prometheus.MustNewConstMetric(m.MinerWorkers, prometheus.CounterValue, float64(len(workerInfos.Infos)))
	for worker, info := range workerInfos.Infos {
		ch <- prometheus.MustNewConstMetric(m.MinerWorkerGPUs, prometheus.CounterValue, float64(info.GPUs), worker)
		ch <- prometheus.MustNewConstMetric(m.MinerWorkerMaintaining, prometheus.CounterValue, float64(info.Maintaining), worker)
		ch <- prometheus.MustNewConstMetric(m.MinerWorkerRejectTask, prometheus.CounterValue, float64(info.RejectTask), worker)
		gpus += info.GPUs
	}
	ch <- prometheus.MustNewConstMetric(m.MinerGPUs, prometheus.CounterValue, float64(gpus))

	checkSectors := m.ml.GetCheckSectors()
	for deadline, sectors := range checkSectors {
		ch <- prometheus.MustNewConstMetric(m.MinerCheckSectorsGood, prometheus.CounterValue, float64(sectors.Good), fmt.Sprintf("%v", deadline))
		ch <- prometheus.MustNewConstMetric(m.MinerCheckSectorsChecked, prometheus.CounterValue, float64(sectors.Checked), fmt.Sprintf("%v", deadline))
	}
	m.mutex.Lock()
	minerId := m.minerInfo.MinerId
	m.mutex.Unlock()

	if 0 < len(minerId) {
		ch <- prometheus.MustNewConstMetric(m.MinerId, prometheus.CounterValue, float64(1), minerId)
		deadlines, err := lotusapi.ProvingDeadlines(m.fullnodeHost, minerId)
		if err == nil {
			for dlIdx, deadline := range deadlines.Deadlines {
				current := 0
				if deadline.Current {
					current = 1
				}
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineAllSectors, prometheus.CounterValue, float64(deadline.AllSectors), fmt.Sprintf("%v", dlIdx))
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineFaultySectors, prometheus.CounterValue, float64(deadline.FaultySectors), fmt.Sprintf("%v", dlIdx))
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineCurrent, prometheus.CounterValue, float64(current), fmt.Sprintf("%v", dlIdx))
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlinePartitions, prometheus.CounterValue, float64(deadline.Partitions), fmt.Sprintf("%v", dlIdx))
				ch <- prometheus.MustNewConstMetric(m.ProvingDeadlineProvenPartitions, prometheus.CounterValue, float64(deadline.ProvenPartitions), fmt.Sprintf("%v", dlIdx))
			}
		} else {
			log.Errorf(log.Fields{}, "fail to get proving deadlines: %v", err)
		}
	}

	filesize := m.ml.LogFileSize()
	ch <- prometheus.MustNewConstMetric(m.LogFileSize, prometheus.CounterValue, float64(filesize))

	chainSyncNotCompletedHosts := m.ml.GetChainSyncNotCompletedHosts()
	//
	for host := range chainSyncNotCompletedHosts {
		ch <- prometheus.MustNewConstMetric(m.ChainSyncNotCompleted, prometheus.CounterValue, float64(1), host)
	}
	chainNotSuitable := m.ml.GetChainNotSuitable()
	ch <- prometheus.MustNewConstMetric(m.ChainNotSuitable, prometheus.CounterValue, float64(chainNotSuitable))

	chainHeadListenSuccessHosts := m.ml.GetChainHeadListenSuccessHosts()
	for host, epoch := range chainHeadListenSuccessHosts {
		ch <- prometheus.MustNewConstMetric(m.ChainHeadListen, prometheus.CounterValue, float64(epoch), host)
	}

	for k, v := range m.storageStat {
		if v != nil {
			ch <- prometheus.MustNewConstMetric(m.StorageMountError, prometheus.CounterValue, 1, k)
		} else {
			ch <- prometheus.MustNewConstMetric(m.StorageMountError, prometheus.CounterValue, 0, k)
		}
		filePerm, _ := systemapi.FilePerm2Int(k)
		ch <- prometheus.MustNewConstMetric(m.StorageMountpointPermission, prometheus.CounterValue, float64(filePerm), k)
	}

	minerFileOpenNumber, _ := systemapi.GetProcessOpenFileNumber("lotus-miner")
	ch <- prometheus.MustNewConstMetric(m.MinerOpenFileNumber, prometheus.CounterValue, float64(minerFileOpenNumber))

	tcpConnectNumber, _ := systemapi.GetProcessTcpConnectNumber("lotus-miner")
	ch <- prometheus.MustNewConstMetric(m.MinerProcessTcpConnectNumber, prometheus.CounterValue, float64(tcpConnectNumber))

	ch <- prometheus.MustNewConstMetric(m.MinerAdjustBaseFee, prometheus.CounterValue, minerAdjustBaseFee)
	ch <- prometheus.MustNewConstMetric(m.MinerAdjustGasFeecap, prometheus.CounterValue, minerAdjustGasFeecap)

	if mineOne.MiningEligible {
		ch <- prometheus.MustNewConstMetric(m.MiningEligible, prometheus.CounterValue, float64(1))
	} else {
		ch <- prometheus.MustNewConstMetric(m.MiningEligible, prometheus.CounterValue, float64(0))
	}
	if mineOne.MiningLateBase {
		ch <- prometheus.MustNewConstMetric(m.MiningLateBase, prometheus.CounterValue, float64(1))
	} else {
		ch <- prometheus.MustNewConstMetric(m.MiningLateBase, prometheus.CounterValue, float64(0))
	}
	if mineOne.MiningLateWinner {
		ch <- prometheus.MustNewConstMetric(m.MiningLateWinner, prometheus.CounterValue, float64(1))
	} else {
		ch <- prometheus.MustNewConstMetric(m.MiningLateWinner, prometheus.CounterValue, float64(0))
	}
	switch mineOne.MiningLateBaseDeltaSecond.(type) {
	case float64:
		ch <- prometheus.MustNewConstMetric(m.MiningLateBaseDeltaSecond, prometheus.CounterValue, mineOne.MiningLateBaseDeltaSecond.(float64))
	case int64:
		ch <- prometheus.MustNewConstMetric(m.MiningLateBaseDeltaSecond, prometheus.CounterValue, float64(mineOne.MiningLateBaseDeltaSecond.(int64)))
	case string:
		miningLateBaseDeltaSecond, _ := strconv.ParseFloat(mineOne.MiningLateBaseDeltaSecond.(string), 64)
		ch <- prometheus.MustNewConstMetric(m.MiningLateBaseDeltaSecond, prometheus.CounterValue, miningLateBaseDeltaSecond)
	}
	miningNetworkPower, _ := strconv.ParseFloat(mineOne.MiningNetworkPower, 64)
	miningMinerPower, _ := strconv.ParseFloat(mineOne.MiningMinerPower, 64)
	ch <- prometheus.MustNewConstMetric(m.MiningNetworkPower, prometheus.CounterValue, miningNetworkPower)
	ch <- prometheus.MustNewConstMetric(m.MiningMinerPower, prometheus.CounterValue, miningMinerPower)

	for _, path := range m.lotusStoragePath {
		pathStatus := getMinerRepoDirUsage(path)
		ch <- prometheus.MustNewConstMetric(m.MinerRepoDirUsage, prometheus.CounterValue, pathStatus.Used, fmt.Sprintf("%v", path), fmt.Sprintf("%v", pathStatus.All))
	}
}

func getMinerRepoDirUsage(dir string) systemapi.DiskStatus {
	return systemapi.DiskUsage(dir)
}
