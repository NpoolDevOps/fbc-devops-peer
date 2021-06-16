package chiaplotterlog

import (
	"strconv"
	"strings"
	"sync"

	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

const (
	RegChiaPlotter = "Total plot creation time was"
)

const (
	KeyChiaPlotter = RegChiaPlotter
)

type LogRegKey struct {
	RegName  string
	ItemName string
}

var logRegKeys = []LogRegKey{
	{
		RegName:  RegChiaPlotter,
		ItemName: KeyChiaPlotter,
	},
}

type ChiaPlotterLog struct {
	logbase         *logbase.Logbase
	newline         chan logbase.LogLine
	chiaPlotterTime float64

	mutex sync.Mutex
}

func NewChiaMinerLog(logfile string) *ChiaPlotterLog {
	newline := make(chan logbase.LogLine)
	cpl := &ChiaPlotterLog{
		logbase: logbase.NewLogbase(logfile, newline),
		newline: newline,
	}

	go cpl.watch()

	return cpl
}

func (cpl *ChiaPlotterLog) setChiaPlotterTime(line logbase.LogLine) {
	ll := line.Msg
	llsec := strings.TrimSpace(strings.Split(strings.TrimSpace(strings.Split(ll, "time was")[1]), "sec")[0])
	llsec2Float, _ := strconv.ParseFloat(llsec, 64)
	cpl.chiaPlotterTime = llsec2Float
}

func (cpl *ChiaPlotterLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		if !cpl.logbase.LineMatchKey(line.Line, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegChiaPlotter:
			cpl.setChiaPlotterTime(line)
		}

	}
}

func (cpl *ChiaPlotterLog) watch() {
	for {
		line := <-cpl.newline
		cpl.processLine(line)
	}
}

func (cpl *ChiaPlotterLog) GetChiaPlotterTime() float64 {
	cpl.mutex.Lock()
	chiaPlotterTime := cpl.chiaPlotterTime
	cpl.mutex.Unlock()
	return chiaPlotterTime
}
