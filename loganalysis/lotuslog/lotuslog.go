package lotuslog

import (
	"strconv"
	"strings"
	"sync"

	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

type LotusLog struct {
	logbase    *logbase.Logbase
	newline    chan logbase.LogLine
	host       string
	hasHost    bool
	largeDelay int64
	tipset     uint64
	spent      uint64

	timeouts uint64
	refuseds uint64
	mutex    sync.Mutex
}

const (
	KeyGotError          = "got error"
	KeyConnectionRefused = "connection refused"
	KeyIOTimeout         = "i/o timeout"
	KeyLargeDelay        = "large delay"
	KeyGatherTipsets     = "\"msg\":\"gathered tipset\","
)

type levelLogKeys struct {
	mainKey string
	subKeys []string
}

var logKeys = []levelLogKeys{
	levelLogKeys{
		mainKey: KeyGotError,
		subKeys: []string{
			KeyConnectionRefused,
			KeyIOTimeout,
		},
	},
	levelLogKeys{
		mainKey: KeyLargeDelay,
		subKeys: []string{},
	},
	levelLogKeys{
		mainKey: KeyGatherTipsets,
		subKeys: []string{},
	},
}

func NewLotusLog(logfile string) *LotusLog {
	newline := make(chan logbase.LogLine)
	ll := &LotusLog{
		logbase: logbase.NewLogbase(logfile, newline),
		newline: newline,
		hasHost: false,
	}

	go ll.watch()

	return ll
}

func (ll *LotusLog) processGotError(line string, subKeys []string) {
	for _, key := range subKeys {
		if !strings.Contains(line, key) {
			continue
		}
		switch key {
		case KeyConnectionRefused:
			ll.mutex.Lock()
			ll.refuseds += 1
			ll.mutex.Unlock()
		case KeyIOTimeout:
			ll.mutex.Lock()
			ll.timeouts += 1
			ll.mutex.Unlock()
		}
	}
}

func (ll *LotusLog) processGatherTipsets(line string) {
	ll.mutex.Lock()
	lstr := strings.Split(line, KeyGatherTipsets)
	llstr := strings.Split(lstr[1], ",")
	tipsets := strings.Split(llstr[0], ":")[1]
	spent := strings.Split(strings.Split(llstr[2], ":")[1], "}")[0]

	ll.tipset, _ = strconv.ParseUint(tipsets, 10, 64)
	ll.spent, _ = strconv.ParseUint(spent, 10, 64)
	ll.mutex.Unlock()

}

func (ll *LotusLog) processLine(line string) {
	for _, key := range logKeys {
		if !strings.Contains(line, key.mainKey) {
			continue
		}

		switch key.mainKey {
		case KeyGotError:
			ll.processGotError(line, key.subKeys)
		case KeyLargeDelay:
			ll.mutex.Lock()
			ll.largeDelay += 1
			ll.mutex.Unlock()
		case KeyGatherTipsets:
			ll.processGatherTipsets(line)
		}
	}
}

func (ll *LotusLog) watch() {
	for {
		line := <-ll.newline
		ll.processLine(line.Line)
	}
}

func (ll *LotusLog) GetRefuseds() uint64 {
	ll.mutex.Lock()
	refuseds := ll.refuseds
	ll.mutex.Unlock()
	return refuseds
}

func (ll *LotusLog) GetTimeouts() uint64 {
	ll.mutex.Lock()
	timeouts := ll.timeouts
	ll.mutex.Unlock()
	return timeouts
}

func (ll *LotusLog) LogFileSize() uint64 {
	return ll.logbase.LogFileSize()
}

func (ll *LotusLog) GetLargeDelay() float64 {
	ll.mutex.Lock()
	largeDelay := ll.largeDelay
	ll.mutex.Unlock()
	return float64(largeDelay)
}

func (ll *LotusLog) GetGatherTipsets() float64 {
	ll.mutex.Lock()
	tipsets := ll.tipset
	ll.mutex.Unlock()
	return float64(tipsets)
}

func (ll *LotusLog) GetTookBlocksSpent() float64 {
	ll.mutex.Lock()
	spent := ll.spent
	ll.mutex.Unlock()
	return float64(spent)
}
