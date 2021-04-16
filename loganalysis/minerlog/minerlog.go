package minerlog

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	lotusapi "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
	"strconv"
	"sync"
	"time"
)

const (
	RegMinedNewBlock     = "\"msg\":\"mined new block\""
	RegMinedPastBlock    = "mined block in the past"
	RegMiningFailedBlock = "\"msg\":\"mining block failed"
	RegRunTaskStart      = "run task start"
	RegRunTaskEnd        = "run task end"
)

const (
	KeyMinedNewBlock     = RegMinedNewBlock
	KeyMinedPastBlock    = RegMinedPastBlock
	KeyMiningFailedBlock = RegMiningFailedBlock
	KeyMinedForkBlock    = "mined fork block"
	KeySectorTask        = "run task"
)

type LogRegKey struct {
	RegName  string
	ItemName string
}

var logRegKeys = []LogRegKey{
	LogRegKey{
		RegName:  RegMinedNewBlock,
		ItemName: KeyMinedNewBlock,
	},
	LogRegKey{
		RegName:  RegMinedPastBlock,
		ItemName: KeyMinedPastBlock,
	},
	LogRegKey{
		RegName:  RegMiningFailedBlock,
		ItemName: KeyMiningFailedBlock,
	},
	LogRegKey{
		RegName:  RegRunTaskStart,
		ItemName: KeySectorTask,
	},
	LogRegKey{
		RegName:  RegRunTaskEnd,
		ItemName: KeySectorTask,
	},
}

type minedBlock struct {
	logbase.LogLine
	Cid     string   `json:"cid"`
	Height  string   `json:"height"`
	Miner   string   `json:"miner"`
	Parents []string `json:"parents"`
	Took    float64  `json:"took"`
}

type sectorTask struct {
	TaskType     string `json:"taskType"`
	taskDone     bool
	TaskStart    uint64 `json:"start"`
	SectorNumber string `json:"sectorNumber"`
	Worker       string `json:"worker"`
	Elapsed      uint64 `json:"elapsed"`
	Error        string `json:"error"`
}

type MinerLog struct {
	logbase         *logbase.Logbase
	newline         chan logbase.LogLine
	items           map[string][]uint64
	fullnodeHost    string
	hasFullnodeHost bool
	candidateBlocks []minedBlock
	forkBlocks      uint64
	pastBlocks      uint64
	failedBlocks    uint64
	sectorTasks     map[string]map[string]sectorTask
	BootTime        uint64
	mutex           sync.Mutex
}

func NewMinerLog(logfile string) *MinerLog {
	newline := make(chan logbase.LogLine)
	ml := &MinerLog{
		logbase:         logbase.NewLogbase(logfile, newline),
		newline:         newline,
		items:           map[string][]uint64{},
		hasFullnodeHost: false,
		sectorTasks:     map[string]map[string]sectorTask{},
		BootTime:        uint64(time.Now().Unix()),
	}

	go ml.watch()

	return ml
}

func (ml *MinerLog) SetFullnodeHost(host string) {
	ml.fullnodeHost = host
	ml.hasFullnodeHost = true
}

func (ml *MinerLog) processMinedNewBlock(line logbase.LogLine) {
	mline := minedBlock{}
	err := json.Unmarshal([]byte(line.Line), &mline)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal %v: %v", line.Line, err)
		return
	}
	ml.candidateBlocks = append(ml.candidateBlocks, mline)
}

func (ml *MinerLog) processSectorTask(line logbase.LogLine, end bool) {
	mline := sectorTask{}
	err := json.Unmarshal([]byte(line.Line), &mline)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal %v: %v", line.Line, err)
	}

	ml.mutex.Lock()
	if _, ok := ml.sectorTasks[mline.TaskType]; !ok {
		ml.sectorTasks[mline.TaskType] = map[string]sectorTask{}
	}
	sectorTasks := ml.sectorTasks[mline.TaskType]
	if end {
		mline.taskDone = true
		if mline.Elapsed == 0 {
			mline.Elapsed = uint64(time.Now().Unix()) - mline.TaskStart
		}
	} else {
		if mline.TaskStart == 0 {
			mline.TaskStart = uint64(time.Now().Unix())
		}
	}
	if len(mline.Worker) == 0 {
		mline.Worker = "miner"
	}
	sectorTasks[mline.SectorNumber] = mline
	ml.sectorTasks[mline.TaskType] = sectorTasks
	ml.mutex.Unlock()
}

func (ml *MinerLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		if !ml.logbase.LineMatchKey(line.Line, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegMinedNewBlock:
			ml.processMinedNewBlock(line)
		case RegMinedPastBlock:
			ml.mutex.Lock()
			ml.pastBlocks += 1
			ml.mutex.Unlock()
		case RegMiningFailedBlock:
			ml.mutex.Lock()
			ml.failedBlocks += 1
			ml.mutex.Unlock()
		case RegRunTaskStart:
			ml.processSectorTask(line, false)
		case RegRunTaskEnd:
			ml.processSectorTask(line, true)
		}

		break
	}
}

func (ml *MinerLog) processCandidateBlocks() {
	if !ml.hasFullnodeHost {
		return
	}

	blocks := []minedBlock{}

	for _, b := range ml.candidateBlocks {
		height, _ := strconv.ParseUint(b.Height, 10, 64)
		cids, err := lotusapi.TipSetByHeight(ml.fullnodeHost, height)
		if err != nil {
			blocks = append(blocks, b)
			continue
		}

		found := false
		for _, cid := range cids {
			if b.Cid == cid {
				found = true
				break
			}
		}

		if !found {
			ml.mutex.Lock()
			ml.forkBlocks += 1
			ml.mutex.Unlock()
			continue
		}

		ml.mutex.Lock()
		blockTimes := ml.items[KeyMinedNewBlock]
		blockTimes = append(blockTimes, uint64(b.Took*1000))
		ml.items[KeyMinedNewBlock] = blockTimes
		ml.mutex.Unlock()
	}

	ml.candidateBlocks = blocks
}

func (ml *MinerLog) watch() {
	for {
		line := <-ml.newline
		ml.processLine(line)
		ml.processCandidateBlocks()
	}
}

func (ml *MinerLog) GetBlockTooks() []uint64 {
	ml.mutex.Lock()
	items := ml.items[KeyMinedNewBlock]
	ml.items[KeyMinedNewBlock] = []uint64{}
	ml.mutex.Unlock()
	return items
}

func (ml *MinerLog) GetForkBlocks() uint64 {
	ml.mutex.Lock()
	forkBlocks := ml.forkBlocks
	ml.forkBlocks = 0
	ml.mutex.Unlock()
	return forkBlocks
}

func (ml *MinerLog) GetPastBlocks() uint64 {
	ml.mutex.Lock()
	pastBlocks := ml.pastBlocks
	ml.pastBlocks = 0
	ml.mutex.Unlock()
	return pastBlocks
}

func (ml *MinerLog) GetFailedBlocks() uint64 {
	ml.mutex.Lock()
	failedBlocks := ml.failedBlocks
	ml.failedBlocks = 0
	ml.mutex.Unlock()
	return failedBlocks
}

type SectorTaskStat struct {
	Worker  string
	Elapsed uint64
	Done    bool
	Sector  string
}

func (ml *MinerLog) GetSectorTasks() map[string]map[string][]SectorTaskStat {
	tasks := map[string]map[string][]SectorTaskStat{}

	ml.mutex.Lock()
	for taskType, sectorTasks := range ml.sectorTasks {
		if _, ok := tasks[taskType]; !ok {
			tasks[taskType] = map[string][]SectorTaskStat{}
		}
		typedTasks := tasks[taskType]
		for _, task := range sectorTasks {
			if _, ok := typedTasks[task.Worker]; !ok {
				typedTasks[task.Worker] = []SectorTaskStat{}
			}
			workerTasks := typedTasks[task.Worker]

			elapsed := task.Elapsed
			if !task.taskDone {
				if 0 < task.TaskStart {
					elapsed = uint64(time.Now().Unix()) - task.TaskStart
				} else {
					elapsed = uint64(time.Now().Unix()) - ml.BootTime
				}
				elapsed = elapsed
			} else {
				delete(ml.sectorTasks[taskType], task.SectorNumber)
			}
			workerTasks = append(workerTasks, SectorTaskStat{
				Worker:  task.Worker,
				Elapsed: elapsed,
				Done:    task.taskDone,
				Sector:  task.SectorNumber,
			})
			typedTasks[task.Worker] = workerTasks
		}
		tasks[taskType] = typedTasks
	}
	ml.mutex.Unlock()

	return tasks
}
