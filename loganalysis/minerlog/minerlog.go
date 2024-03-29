package minerlog

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	lotusapi "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

const (
	RegMinedNewBlock         = "\"msg\":\"mined new block\""
	RegMinedPastBlock        = "mined block in the past"
	RegMiningFailedBlock     = "\"msg\":\"mining block failed"
	RegRunTaskStart          = "run task start"
	RegRunTaskEnd            = "run task end"
	RegCheckSectors          = "\"msg\":\"Checked sectors\""
	RegChainSyncNotCompleted = "chain sync state is not completed of "
	RegChainNotSuitable      = "cannot find suitable fullnode"
	RegChainHeadListen       = "success to listen chain head from "
	RegGetFeeMiner           = "adjust fee for nonce"
	RegMinerIsMaster         = "play as master"
	RegMineOne               = "\"msg\":\"completed mineOne\""
	RegComputingWindowPost   = "\"msg\":\"computing window post\""
)

const (
	KeyMinedNewBlock         = RegMinedNewBlock
	KeyMinedPastBlock        = RegMinedPastBlock
	KeyMiningFailedBlock     = RegMiningFailedBlock
	KeyMinedForkBlock        = "mined fork block"
	KeySectorTask            = "run task"
	KeyCheckSectors          = RegCheckSectors
	KeyChainSyncNotCompleted = RegChainSyncNotCompleted
	KeyChainNotSuitable      = RegChainNotSuitable
	KeyChainHeadListen       = RegChainHeadListen
	keyGetFeeMiner           = RegGetFeeMiner
	keyMinerIsMaster         = RegMinerIsMaster
	keyMineOne               = RegMineOne
	keyComputingWindowPost   = RegComputingWindowPost
)

type LogRegKey struct {
	RegName  string
	ItemName string
}

var logRegKeys = []LogRegKey{
	{
		RegName:  RegMinedNewBlock,
		ItemName: KeyMinedNewBlock,
	},
	{
		RegName:  RegMinedPastBlock,
		ItemName: KeyMinedPastBlock,
	},
	{
		RegName:  RegMiningFailedBlock,
		ItemName: KeyMiningFailedBlock,
	},
	{
		RegName:  RegRunTaskStart,
		ItemName: KeySectorTask,
	},
	{
		RegName:  RegRunTaskEnd,
		ItemName: KeySectorTask,
	},
	{
		RegName:  RegCheckSectors,
		ItemName: KeyCheckSectors,
	},
	{
		RegName:  RegChainSyncNotCompleted,
		ItemName: KeyChainSyncNotCompleted,
	},
	{
		RegName:  RegChainNotSuitable,
		ItemName: KeyChainNotSuitable,
	},
	{
		RegName:  RegChainHeadListen,
		ItemName: KeyChainHeadListen,
	},
	{
		RegName:  RegGetFeeMiner,
		ItemName: keyGetFeeMiner,
	},
	{
		RegName:  RegMinerIsMaster,
		ItemName: keyMinerIsMaster,
	},
	{
		RegName:  RegMineOne,
		ItemName: keyMineOne,
	},
	{
		RegName:  RegComputingWindowPost,
		ItemName: keyComputingWindowPost,
	},
}

type minedBlock struct {
	logbase.LogLine
	Cid       string      `json:"cid"`
	Height    interface{} `json:"height"`
	Miner     string      `json:"miner"`
	Parents   []string    `json:"parents"`
	Took      float64     `json:"took"`
	InThePast bool
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

type CheckSectors struct {
	Good     int `json:"good"`
	Checked  int `json:"checked"`
	Deadline int `json:"deadline"`
}

type MineOne struct {
	MiningLateBase            bool        `json:"lateStart"`
	MiningLateWinner          bool        `json:"isWinner"`
	MiningLateBaseDeltaSecond interface{} `json:"baseDeltaSeconds"`
	MiningEligible            bool        `json:"isEligible"`
	MiningNetworkPower        string      `json:"networkPowerAtLookback"`
	MiningMinerPower          string      `json:"minerPowerAtLookback"`
}

type ComputingWindowPost struct {
	Batch    uint    `json:"batch"`
	Elapsed  float64 `json:"elapsed"`
	Deadline uint64  `json:"deadline"`
}

type MinerLog struct {
	logbase                    *logbase.Logbase
	newline                    chan logbase.LogLine
	items                      map[string][]uint64
	fullnodeHost               string
	hasFullnodeHost            bool
	candidateBlocks            []minedBlock
	forkBlocks                 uint64
	pastBlocks                 uint64
	failedBlocks               uint64
	sectorTasks                map[string]map[string]sectorTask
	BootTime                   uint64
	checkSectors               map[int]CheckSectors
	chainSyncNotCompletedHosts map[string]struct{}
	chainNotSuitable           uint64
	chainHeadListenHosts       map[string]uint64
	minerAdjustGasFeecap       float64
	minerAdjustBaseFee         float64
	minerIsMaster              bool
	mineOne                    MineOne
	computingWindowPostGroup   []ComputingWindowPost
	timeStamp                  uint64
	sectorGroup                []sectorTask
	mutex                      sync.Mutex
}

func (ml *MinerLog) processComputingWindowPost(line logbase.LogLine) {
	computingWindowPost := ComputingWindowPost{}
	err := json.Unmarshal([]byte(line.Line), &computingWindowPost)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal %v: %v", line.Line, err)
		return
	}
	ml.mutex.Lock()
	has := false
	for index, provingGroup := range ml.computingWindowPostGroup {
		if provingGroup.Batch == computingWindowPost.Batch && provingGroup.Deadline == computingWindowPost.Deadline {
			ml.computingWindowPostGroup[index].Elapsed = computingWindowPost.Elapsed
			has = true
		}
	}
	if !has {
		ml.computingWindowPostGroup = append(ml.computingWindowPostGroup, computingWindowPost)
	}
	ml.mutex.Unlock()
}

func (ml *MinerLog) processMineOne(line logbase.LogLine) {
	mineOne := MineOne{}
	err := json.Unmarshal([]byte(line.Line), &mineOne)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal %v: %v", line.Line, err)
		return
	}
	ml.mutex.Lock()
	ml.mineOne = mineOne
	ml.mutex.Unlock()
}

func (ml *MinerLog) processMinerIsMaster(line logbase.LogLine) {
	ll := line.Msg
	isMasterStr := strings.TrimSpace(strings.Split(ll, RegMinerIsMaster)[1])
	if isMasterStr == "true" {
		ml.minerIsMaster = true
	} else {
		ml.minerIsMaster = false
	}

}

func (ml *MinerLog) setMinerFee(line logbase.LogLine) {
	ll := line.Msg
	llarr := strings.Split(ll, "feecap ->")
	minerAdjustGasFeecap := strings.TrimSpace(strings.Split(llarr[1], "|")[0])
	minerAdjustBaseFee := strings.TrimSpace(strings.Split(llarr[1], "|")[1])

	minerAdjustBaseFee2Float, _ := strconv.ParseFloat(minerAdjustBaseFee, 64)
	minerAdjustGasFeecap2Float, _ := strconv.ParseFloat(minerAdjustGasFeecap, 64)

	ml.minerAdjustGasFeecap = minerAdjustGasFeecap2Float
	ml.minerAdjustBaseFee = minerAdjustBaseFee2Float
}

func NewMinerLog(logfile string) *MinerLog {
	newline := make(chan logbase.LogLine)
	ml := &MinerLog{
		logbase:                    logbase.NewLogbase(logfile, newline),
		newline:                    newline,
		items:                      map[string][]uint64{},
		hasFullnodeHost:            false,
		sectorTasks:                map[string]map[string]sectorTask{},
		BootTime:                   uint64(time.Now().Unix()),
		checkSectors:               map[int]CheckSectors{},
		chainSyncNotCompletedHosts: map[string]struct{}{},
		chainHeadListenHosts:       map[string]uint64{},
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

func (ml *MinerLog) processMinedInPastBlock(line logbase.LogLine) {
	mline := minedBlock{}
	err := json.Unmarshal([]byte(line.Line), &mline)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal %v: %v", line.Line, err)
		return
	}

	for i, b := range ml.candidateBlocks {
		if b.Cid == mline.Cid {
			ml.candidateBlocks[i].InThePast = true
			break
		}
	}
}

func (ml *MinerLog) processSectorTask(line logbase.LogLine, end bool) {
	mline := sectorTask{}
	err := json.Unmarshal([]byte(line.Line), &mline)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal %v: %v", line.Line, err)
		return
	}

	ml.mutex.Lock()
	if _, ok := ml.sectorTasks[mline.TaskType]; !ok {
		ml.sectorTasks[mline.TaskType] = map[string]sectorTask{}
	}
	sectorTasks := ml.sectorTasks[mline.TaskType]
	if end {
		mline.taskDone = true
		if mline.Elapsed == 0 {
			mline.Elapsed = ml.timeStamp - mline.TaskStart
		}
	} else {
		if mline.TaskStart == 0 {
			mline.TaskStart = ml.timeStamp
		}
	}
	if len(mline.Worker) == 0 {
		mline.Worker = "miner"
	}

	finish := false
	for _, sector := range ml.sectorGroup {
		if sector.SectorNumber == mline.SectorNumber && sector.TaskType == mline.TaskType {
			finish = true
		}
	}
	if !finish {
		ml.sectorGroup = append(ml.sectorGroup, mline)
	}
	sectorTasks[mline.SectorNumber] = mline
	ml.sectorTasks[mline.TaskType] = sectorTasks

	for {
		if len(ml.sectorGroup) <= 1000 {
			break
		}
		delete(ml.sectorTasks[ml.sectorGroup[0].TaskType], ml.sectorGroup[0].SectorNumber)
		ml.sectorGroup = ml.sectorGroup[1:]
	}

	ml.mutex.Unlock()
}

func (ml *MinerLog) processCheckSectors(line logbase.LogLine) {
	cs := CheckSectors{
		Deadline: -1,
	}
	err := json.Unmarshal([]byte(line.Line), &cs)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse %v to check sectors: %v", line.Line, err)
		return
	}

	if cs.Deadline < 0 {
		cs.Deadline = rand.Int()
	}

	ml.mutex.Lock()
	ml.checkSectors[cs.Deadline] = cs
	ml.mutex.Unlock()
}

func (ml *MinerLog) processChainSyncNotCompleted(line logbase.LogLine) {
	msg := strings.Replace(line.Msg, RegChainSyncNotCompleted, "", -1)

	host := ""
	if strings.HasPrefix(msg, "ws://") {
		ss := strings.Split(msg, "ws://")
		if len(ss) < 2 {
			log.Errorf(log.Fields{}, "cannot parse line: %v", line.Msg)
			return
		}
		ss = strings.Split(ss[1], ":")
		if len(ss) < 2 {
			log.Errorf(log.Fields{}, "cannot parse line: %v", line.Msg)
			return
		}
		host = ss[0]
	} else if strings.HasPrefix(msg, "mainnode") {
		host = "mainnode"
	}
	ml.chainSyncNotCompletedHosts[host] = struct{}{}
}

func (ml *MinerLog) processChainHeadListen(line logbase.LogLine) {
	msg := strings.Replace(line.Msg, RegChainHeadListen, "", -1)

	host := ""
	if strings.HasPrefix(msg, "ws://") {
		ss := strings.Split(msg, "ws://")
		if len(ss) < 2 {
			log.Errorf(log.Fields{}, "cannot parse line: %v", line.Msg)
			return
		}
		ss = strings.Split(ss[1], ":")
		if len(ss) < 2 {
			log.Errorf(log.Fields{}, "cannot parse line: %v", line.Msg)
			return
		}
		host = ss[0]
	} else if strings.HasPrefix(msg, "mainnode") {
		host = "mainnode"
	}
	curepoch := strings.Split(msg, " ")[2]
	epoch, _ := strconv.ParseInt(curepoch, 10, 64)
	ml.chainHeadListenHosts[host] = uint64(epoch)
}

type Line struct {
	TimeStamp string `json:"ts"`
}

func (ml *MinerLog) setTimeStamp(line logbase.LogLine) {
	var everyLine Line
	ml.mutex.Lock()
	err := json.Unmarshal([]byte(line.Line), &everyLine)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to unmarshal &v, err is %v", line.Line, err)
	} else {
		theTime, _ := time.Parse("2006-01-02T15:04:05.000", strings.TrimSpace(strings.Split(everyLine.TimeStamp, "+")[0]))
		ml.timeStamp = uint64(theTime.Unix() - 28800)
	}
	ml.mutex.Unlock()
}

func (ml *MinerLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		ml.setTimeStamp(line)
		if !ml.logbase.LineMatchKey(line.Line, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegMinedNewBlock:
			ml.processMinedNewBlock(line)
		case RegMinedPastBlock:
			ml.mutex.Lock()
			ml.pastBlocks += 1
			ml.processMinedInPastBlock(line)
			ml.mutex.Unlock()
		case RegMiningFailedBlock:
			ml.mutex.Lock()
			ml.failedBlocks += 1
			ml.mutex.Unlock()
		case RegRunTaskStart:
			ml.processSectorTask(line, false)
		case RegRunTaskEnd:
			ml.processSectorTask(line, true)
		case RegCheckSectors:
			ml.processCheckSectors(line)
		case RegChainSyncNotCompleted:
			ml.processChainSyncNotCompleted(line)
		case RegChainNotSuitable:
			ml.mutex.Lock()
			ml.chainNotSuitable = 1
			ml.mutex.Unlock()
		case RegChainHeadListen:
			ml.processChainHeadListen(line)
		case RegGetFeeMiner:
			ml.setMinerFee(line)
		case RegMinerIsMaster:
			ml.processMinerIsMaster(line)
		case RegMineOne:
			ml.processMineOne(line)
		case RegComputingWindowPost:
			ml.processComputingWindowPost(line)
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
		if b.InThePast {
			continue
		}

		var height uint64

		switch b.Height.(type) {
		case float64:
			height = uint64(b.Height.(float64))
		case int:
			height = uint64(b.Height.(int))
		case int64:
			height = uint64(b.Height.(int64))
		case string:
			height, _ = strconv.ParseUint(b.Height.(string), 10, 64)
		}

		cids, err := lotusapi.TipSetByHeight(ml.fullnodeHost, height)
		if err != nil {
			blocks = append(blocks, b)
			continue
		}

		chainHeight, _ := lotusapi.ChainHeadHeight(ml.fullnodeHost)

		if height > uint64(chainHeight) {
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
	ml.mutex.Unlock()
	return items
}

func (ml *MinerLog) GetForkBlocks() uint64 {
	ml.mutex.Lock()
	forkBlocks := ml.forkBlocks
	ml.mutex.Unlock()
	return forkBlocks
}

func (ml *MinerLog) GetPastBlocks() uint64 {
	ml.mutex.Lock()
	pastBlocks := ml.pastBlocks
	ml.mutex.Unlock()
	return pastBlocks
}

func (ml *MinerLog) GetFailedBlocks() uint64 {
	ml.mutex.Lock()
	failedBlocks := ml.failedBlocks
	ml.mutex.Unlock()
	return failedBlocks
}

func (ml *MinerLog) GetChainSyncNotCompletedHosts() map[string]struct{} {
	ml.mutex.Lock()
	hosts := ml.chainSyncNotCompletedHosts
	ml.mutex.Unlock()
	return hosts
}

func (ml *MinerLog) GetChainNotSuitable() uint64 {
	ml.mutex.Lock()
	chainNotSuitable := ml.chainNotSuitable
	ml.mutex.Unlock()
	return chainNotSuitable
}

func (ml *MinerLog) GetChainHeadListenSuccessHosts() map[string]uint64 {
	ml.mutex.Lock()
	hosts := ml.chainHeadListenHosts
	ml.mutex.Unlock()
	return hosts
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
					elapsed = ml.timeStamp - task.TaskStart
				} else {
					elapsed = ml.timeStamp - ml.BootTime
				}
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

func (ml *MinerLog) GetCheckSectors() map[int]CheckSectors {
	ml.mutex.Lock()
	sectors := ml.checkSectors
	ml.mutex.Unlock()
	return sectors
}

func (ml *MinerLog) LogFileSize() uint64 {
	return ml.logbase.LogFileSize()
}

func (ml *MinerLog) GetMinerFeeAdjustGasFeecap() float64 {
	ml.mutex.Lock()
	minerAdjustGasFeecap := ml.minerAdjustGasFeecap
	ml.mutex.Unlock()
	return minerAdjustGasFeecap
}

func (ml *MinerLog) GetMinerAdjustBaseFee() float64 {
	ml.mutex.Lock()
	minerAdjustBaseFee := ml.minerAdjustBaseFee
	ml.mutex.Unlock()
	return minerAdjustBaseFee
}

func (ml *MinerLog) GetMinerIsMaster() float64 {
	var minerIsMaster float64
	ml.mutex.Lock()
	if ml.minerIsMaster {
		minerIsMaster = 1
	} else {
		minerIsMaster = 0
	}
	ml.mutex.Unlock()
	return minerIsMaster
}

func (ml *MinerLog) GetMineOne() MineOne {
	ml.mutex.Lock()
	mineOne := ml.mineOne
	ml.mutex.Unlock()
	return mineOne
}

func (ml *MinerLog) GetWindowPostProving() []ComputingWindowPost {
	ml.mutex.Lock()
	computingwindowPostGroup := ml.computingWindowPostGroup
	ml.mutex.Unlock()
	return computingwindowPostGroup
}
