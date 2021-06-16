package chiaminerlog

import (
	"strings"
	"sync"

	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

const (
	RegChiaMinerTimeout = "msg"
)

const (
	KeyChiaMinerTimeout = RegChiaMinerTimeout
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
}

type ChiaMinerLog struct {
	logbase          *logbase.Logbase
	newline          chan logbase.LogLine
	chiaMinerTimeout bool

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

func (cml *ChiaMinerLog) setChiaMinerTimeout(line logbase.LogLine) {
	ll := line.Msg
	result := strings.TrimSpace(strings.Split(strings.TrimSpace(strings.Split(ll, "msg=")[0]), "level=")[1])
	if result == "error" {
		cml.chiaMinerTimeout = true
	} else {
		cml.chiaMinerTimeout = false
	}
}

func (cml *ChiaMinerLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		if !cml.logbase.LineMatchKey(line.Line, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegChiaMinerTimeout:
			cml.setChiaMinerTimeout(line)
		}
	}
}

func (cml *ChiaMinerLog) watch() {
	for {
		line := <-cml.newline
		cml.processLine(line)
	}
}

func (cml *ChiaMinerLog) GetChiaMinerTimeout() bool {
	cml.mutex.Lock()
	chiaMinerTimeout := cml.chiaMinerTimeout
	cml.mutex.Unlock()
	return chiaMinerTimeout
}
