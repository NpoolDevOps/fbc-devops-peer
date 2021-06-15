package lotuslog

import (
	"strings"
	"sync"

	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

type LotusLog struct {
	logbase *logbase.Logbase
	newline chan logbase.LogLine
	host    string
	hasHost bool

	timeouts uint64
	refuseds uint64
	mutex    sync.Mutex
}

const (
	KeyGotError          = "got error"
	KeyConnectionRefused = "connection refused"
	KeyIOTimeout         = "i/o timeout"
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

func (ll *LotusLog) processLine(line string) {
	for _, key := range logKeys {
		if !strings.Contains(line, key.mainKey) {
			continue
		}

		switch key.mainKey {
		case KeyGotError:
			ll.processGotError(line, key.subKeys)
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
