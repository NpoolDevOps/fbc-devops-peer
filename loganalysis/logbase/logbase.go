package logbase

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LogLine struct {
	Level     string `json:"level"`
	Logger    string `json:"logger"`
	Caller    string `json:"caller"`
	Timestamp string `json:"ts"`
	Msg       string `json:"msg"`
	Line      string
}

func (ll *LogLine) String() string {
	return fmt.Sprintf("%v	%v", ll.Timestamp, ll.Msg)
}

type Logbase struct {
	tail        *tail.Tail
	newline     chan LogLine
	lastLogTime uint64
	logfile     string
	logTsFile   string
	logTsPath   string
}

func NewLogbase(logfile string, newline chan LogLine) *Logbase {
	lb := &Logbase{
		newline:   newline,
		logfile:   logfile,
		logTsFile: fmt.Sprintf(".%v.timestamp", path.Base(logfile)),
		logTsPath: filepath.Join(os.Getenv("HOME"), ".fbc-devios-peer"),
	}
	lb.tail, _ = tail.TailFile(logfile, tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: false,
	})

	b, err := ioutil.ReadFile(filepath.Join(lb.logTsPath, lb.logTsFile))
	if err == nil {
		lb.lastLogTime = lb.Timestamp(string(b))
	}

	go lb.watch()

	return lb
}

func (lb *Logbase) parseTimestamp(ts string) uint64 {
	out, _ := exec.Command("date", "-d", ts, "+%s").Output()
	t, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	return t
}

func (lb *Logbase) watch() {
	for {
		line, ok := <-lb.tail.Lines
		if !ok {
			time.Sleep(1 * time.Second)
			continue
		}
		logLine := LogLine{}
		err := json.Unmarshal([]byte(line.Text), &logLine)
		if err == nil {
			timestamp := lb.Timestamp(logLine.Timestamp)
			if timestamp <= lb.lastLogTime {
				continue
			}

			logLine.Line = line.Text
			lb.newline <- logLine

			os.MkdirAll(lb.logTsPath, 0666)
			err = ioutil.WriteFile(filepath.Join(lb.logTsPath, lb.logTsFile),
				[]byte(logLine.Timestamp), 0666)
			if err != nil {
				log.Errorf(log.Fields{}, "cannot write timestamp: %v", err)
			}
		}
	}
}

func (lb *Logbase) Timestamp(line string) uint64 {
	return lb.parseTimestamp(line)
}

func (lb *Logbase) LineMatchKey(line string, key string) bool {
	return strings.Contains(line, key)
}
