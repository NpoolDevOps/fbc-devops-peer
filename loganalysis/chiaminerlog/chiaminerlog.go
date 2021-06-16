package chiaminerlog

import (
	"sync"

	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

const (
	RegChiaMinerTimeout      = "扫盘超时"
	RegChiaMinerTimeoutLimit = "超过最大时间限制"
)

const (
	KeyChiaMinerTimeout      = RegChiaMinerTimeout
	KeyChiaMinerTimeoutLimit = RegChiaMinerTimeoutLimit
)

type LogRegKey struct {
	RegName  string
	ItemName string
}

var logRegKeys = []LogRegKey{
	{
		RegName:  RegChiaMinerTimeout,
		ItemName: KeyChiaMinerTimeout,
	},
	{
		RegName:  RegChiaMinerTimeoutLimit,
		ItemName: KeyChiaMinerTimeoutLimit,
	},
}

type ChiaMinerLog struct {
	logbase          *logbase.Logbase
	newline          chan logbase.LogLine
	chiaMinerTimeout uint64

	mutex sync.Mutex
}

func NewChiaMinerLog(logfile string) *ChiaMinerLog {
	newline := make(chan logbase.LogLine)
	cml := &ChiaMinerLog{
		logbase: logbase.NewLogbase(logfile, newline),
		newline: newline,
	}

	go cml.watch()

	return cml
}

func (cml *ChiaMinerLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		if !cml.logbase.LineMatchKey(line.Line, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegChiaMinerTimeout:
			cml.mutex.Lock()
			cml.chiaMinerTimeout += 1
			cml.mutex.Unlock()
		case RegChiaMinerTimeoutLimit:
			cml.mutex.Lock()
			cml.chiaMinerTimeout += 1
			cml.mutex.Unlock()
		}

	}
}

func (cml *ChiaMinerLog) watch() {
	for {
		line := <-cml.newline
		cml.processLine(line)
	}
}

func (cml *ChiaMinerLog) GetChiaMinerTimeout() uint64 {
	cml.mutex.Lock()
	chiaMinerTimeout := cml.chiaMinerTimeout
	cml.mutex.Unlock()
	return chiaMinerTimeout
}
