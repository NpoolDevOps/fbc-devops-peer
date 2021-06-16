package chiaplotterlog

import (
	"strconv"
	"strings"
	"sync"

	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/logbase"
)

const (
	RegPlotterPlottingTime = "Total plot creation time was"
)

const (
	KeyPlotterPlottingTime = RegPlotterPlottingTime
)

type LogRegKey struct {
	RegName  string
	ItemName string
}

var logRegKeys = []LogRegKey{
	{
		RegName:  RegPlotterPlottingTime,
		ItemName: KeyPlotterPlottingTime,
	},
}

type ChiaPlotterLog struct {
	logbase           *logbase.Logbase
	newline           chan logbase.LogLine
	chiaPlotMaxTime   float64
	chiaPlotTotalTime float64
	chiaPlotMinTime   float64
	chiaPlotCount     int64

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

func (cpl *ChiaPlotterLog) parseChiaPlotterTime(line logbase.LogLine) {
	ll := line.Msg
	llsec := strings.TrimSpace(strings.Split(strings.TrimSpace(strings.Split(ll, "time was")[1]), "sec")[0])
	llsec2Float, _ := strconv.ParseFloat(llsec, 64)

	cpl.mutex.Lock()
	if llsec2Float >= cpl.chiaPlotMaxTime || cpl.chiaPlotMaxTime == 0 {
		cpl.chiaPlotMaxTime = llsec2Float
	}
	if cpl.chiaPlotMinTime >= llsec2Float || cpl.chiaPlotMinTime == 0 {
		cpl.chiaPlotMinTime = llsec2Float
	}

	cpl.chiaPlotTotalTime += llsec2Float
	cpl.chiaPlotCount++
	cpl.mutex.Unlock()
}

func (cpl *ChiaPlotterLog) processLine(line logbase.LogLine) {
	for _, item := range logRegKeys {
		if !cpl.logbase.LineMatchKey(line.Line, item.RegName) {
			continue
		}

		switch item.RegName {
		case RegPlotterPlottingTime:
			cpl.parseChiaPlotterTime(line)
		}

	}
}

func (cpl *ChiaPlotterLog) watch() {
	for {
		line := <-cpl.newline
		cpl.processLine(line)
	}
}

func (cpl *ChiaPlotterLog) GetChiaPlotterMaxTime() float64 {
	cpl.mutex.Lock()
	chiaPlotterMaxTime := cpl.chiaPlotMaxTime
	cpl.mutex.Unlock()
	return chiaPlotterMaxTime
}

func (cpl *ChiaPlotterLog) GetChiaPlotterAvgTime() float64 {
	cpl.mutex.Lock()
	chiaPlotterAvgTime := cpl.chiaPlotTotalTime / float64(cpl.chiaPlotCount)
	cpl.mutex.Unlock()
	return chiaPlotterAvgTime
}

func (cpl *ChiaPlotterLog) GetChiaPlotterMinTime() float64 {
	cpl.mutex.Lock()
	chiaPlotterMinTime := cpl.chiaPlotMinTime
	cpl.mutex.Unlock()
	return chiaPlotterMinTime
}

func (cpl *ChiaPlotterLog) GetParseChiaPlotterTimeCount() int64 {
	cpl.mutex.Lock()
	parseChiaPlotterTimeCount := cpl.chiaPlotCount
	cpl.mutex.Unlock()
	return parseChiaPlotterTimeCount
}
