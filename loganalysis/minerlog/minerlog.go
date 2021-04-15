package minerlog

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	lotusapi "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
	"strconv"
	"sync"
)

const (
	RegMinedNewBlock = "mined new block"
	RegRunTaskStart  = "run task start"
	RegRunTaskEnd    = "run task end"
)

const (
	KeyMinedNewBlock  = RegMinedNewBlock
	KeyMinedForkBlock = "mined fork block"
	KeySectorTask     = "run task"
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

type MinerLog struct {
	logbase         *logbase.Logbase
	newline         chan logbase.LogLine
	items           map[string][]uint64
	fullnodeHost    string
	hasFullnodeHost bool
	candidateBlocks []minedBlock
	forkBlocks      uint64
	mutex           sync.Mutex
}

func NewMinerLog(logfile string) *MinerLog {
	newline := make(chan logbase.LogLine)
	ml := &MinerLog{
		logbase:         logbase.NewLogbase(logfile, newline),
		newline:         newline,
		items:           map[string][]uint64{},
		hasFullnodeHost: false,
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
	}
	ml.candidateBlocks = append(ml.candidateBlocks, mline)
}

func (ml *MinerLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		if !ml.logbase.LineMatchKey(line.Msg, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegMinedNewBlock:
			ml.processMinedNewBlock(line)
		}

		break
	}
}

func (ml *MinerLog) processCandidateBlocks() {
	if !ml.hasFullnodeHost {
		return
	}

	for _, b := range ml.candidateBlocks {
		height, _ := strconv.ParseUint(b.Height, 10, 64)
		cids, err := lotusapi.TipSetByHeight(ml.fullnodeHost, height)
		if err != nil {
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
	ml.mutex.Unlock()
	return forkBlocks
}