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
		cml.chiaMinerTimeout = false
	} else {
		cml.chiaMinerTimeout = true
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

// time="2021-06-10T11:30:47Z" level=info msg="new mining info" capacity="0 B" file=loggers.go func=logging.CPrint height=411126 jobId=999051335 line=168 scan consume=15000 scan time="2021-06-10 11:30:38" tid=2482550
// time="2021-06-10T11:30:47Z" level=error msg="您的扫盘时间(15.00s)超过最大时间限制(10.00s)" f7="{loggers.go,logging.CPrint,156}" f8="{wrapper_miner.go,miner.(*Wrapper).Start.func3,125}" f9="{server.go,server.(*Server).MiningInfo,36}" tid=2482550
// time="2021-06-10T11:30:58Z" level=error msg="扫盘超时" f7="{loggers.go,logging.CPrint,156}" f8="{miner.go,chia.(*Miner).MinerStatus,247}" f9="{wrapper_miner.go,miner.(*Wrapper).Start.func3,117}" path=/mnt/sda1/gva/chia-nm-forrest-2/ scan limit=10000 scan time=15000 tid=2482550
// time="2021-06-10T11:30:58Z" level=error msg="扫盘超时" f7="{loggers.go,logging.CPrint,156}" f8="{miner.go,chia.(*Miner).MinerStatus,247}" f9="{wrapper_miner.go,miner.(*Wrapper).Start.func3,117}" path=/mnt/sdb1/gvb/chia-nm-forrest-2/ scan limit=10000 scan time=15000 tid=2482550
