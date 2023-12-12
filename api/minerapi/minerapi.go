package minerapi

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
)

type MinerInfo struct {
	MinerId    string
	SectorSize float64
	Power      float64
	Raw        float64
	Committed  float64
	Proving    float64
	Faulty     float64

	MinerBalance     float64
	InitialPledge    float64
	PrecommitDeposit float64
	Vesting          float64
	Available        float64

	WorkerBalance  float64
	ControlBalance float64

	State map[string]uint64
}

func parseBalance(line string) float64 {
	balance := strings.Split(line, ":")[1]
	balance = strings.TrimSpace(balance)
	balance = strings.Split(balance, " ")[0]
	b, _ := strconv.ParseFloat(balance, 64)
	return b
}

func GetMinerInfo(ch chan MinerInfo, sectors bool) {
	go func() {
		inSectorState := false
		hideSector := "--hide-sectors-info=true"
		if sectors {
			hideSector = "--hide-sectors-info=false"
		}

		info := MinerInfo{
			State: map[string]uint64{},
		}

		out, err := systemapi.RunCommand(exec.Command("/usr/local/bin/lotus-miner", "info", hideSector))
		if err != nil {
			log.Errorf(log.Fields{}, "fail to run lotus-miner info: %v", err)
			ch <- info
			return
		}
		br := bufio.NewReader(bytes.NewReader(out))

		for {
			line, _, err := br.ReadLine()
			if err != nil {
				break
			}

			lineStr := strings.TrimSpace(string(line))
			if lineStr == "" {
				continue
			}

			if strings.Contains(lineStr, "Miner: ") {
				info.MinerId = strings.Split(lineStr, " ")[1]
				sectorSize := strings.Split(lineStr, "(")[1]
				sectorSize = strings.Split(sectorSize, ")")[0]
				sectorSize = strings.TrimSpace(strings.Split(sectorSize, "GiB sectors")[0])
				info.SectorSize, _ = strconv.ParseFloat(sectorSize, 64)
			}
			if strings.Contains(lineStr, "Power: ") {
				info.Power, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
				info.Power = convertTiB(info.Power, lineStr)
			}
			if strings.Contains(lineStr, "Raw: ") {
				info.Raw, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
				info.Raw = convertTiB(info.Raw, lineStr)
			}
			if strings.Contains(lineStr, "Committed: ") {
				info.Committed, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
				info.Committed = convertTiB(info.Committed, lineStr)
			}
			if !inSectorState {
				if strings.Contains(lineStr, "Proving: ") {
					info.Proving, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
					info.Proving = convertTiB(info.Proving, lineStr)
					if strings.Contains(lineStr, "Faulty, ") {
						faulty := strings.Split(lineStr, "(")[1]
						info.Faulty, _ = strconv.ParseFloat(strings.Split(faulty, " ")[0], 64)
						info.Faulty = convertTiB(info.Faulty, lineStr)
					}
				}
			}
			if strings.Contains(lineStr, "Miner Balance: ") {
				info.MinerBalance = parseBalance(lineStr)
			}
			if strings.Contains(lineStr, "PreCommit: ") {
				info.PrecommitDeposit = parseBalance(lineStr)
			}
			if strings.Contains(lineStr, "Pledge: ") {
				info.InitialPledge = parseBalance(lineStr)
			}
			if strings.Contains(lineStr, "Vesting: ") {
				info.Vesting = parseBalance(lineStr)
			}
			if strings.Contains(lineStr, "Available: ") && info.Available == 0 {
				info.Available = parseBalance(lineStr)
			}
			if strings.Contains(lineStr, "Worker Balance: ") {
				info.WorkerBalance = parseBalance(lineStr)
			}
			if strings.Contains(lineStr, "Control: ") {
				info.ControlBalance = parseBalance(lineStr)
			}
			if inSectorState {
				state := strings.Split(lineStr, ": ")[0]
				count := strings.Split(lineStr, " ")[1]
				info.State[state], _ = strconv.ParseUint(count, 10, 64)
			}
			if strings.Contains(lineStr, "Sectors:") {
				inSectorState = true
			}
		}

		ch <- info
	}()
}

type SealingJob struct {
	Running    uint64
	Assigned   uint64
	MaxWaiting uint64
	MaxRunning uint64
}

type SealingJobs struct {
	Jobs map[string]map[string]SealingJob
}

func GetSealingJobs(ch chan SealingJobs) {
	go func() {
		info := SealingJobs{
			Jobs: map[string]map[string]SealingJob{},
		}

		out, err := systemapi.RunCommand(exec.Command("/usr/local/bin/lotus-miner", "sealing", "jobs"))
		if err != nil {
			log.Errorf(log.Fields{}, "fail to run lotus-miner sealing jobs: %v", err)
			ch <- info
			return
		}

		br := bufio.NewReader(bytes.NewReader(out))

		titleLine := true

		for {
			line, _, err := br.ReadLine()
			if err != nil {
				break
			}

			if titleLine {
				titleLine = false
				continue
			}

			lineStr := strings.TrimSpace(string(line))
			items := strings.Fields(lineStr)
			if _, ok := info.Jobs[items[4]]; !ok {
				info.Jobs[items[4]] = map[string]SealingJob{}
			}
			jobs := info.Jobs[items[4]]
			if _, ok := jobs[items[3]]; !ok {
				jobs[items[3]] = SealingJob{}
			}
			job := jobs[items[3]]

			elapsedDuration, _ := time.ParseDuration(items[6])
			elapsed := uint64(elapsedDuration.Seconds())

			switch items[5] {
			case "running":
				job.Running += 1
				if job.MaxRunning < elapsed {
					job.MaxRunning = elapsed
				}
			default:
				job.Assigned += 1
				if job.MaxWaiting < elapsed {
					job.MaxWaiting = elapsed
				}
			}

			jobs[items[3]] = job
			info.Jobs[items[4]] = jobs
		}

		ch <- info
	}()
}

type WorkerInfo struct {
	GPUs        int
	Maintaining int
	RejectTask  int
}

type WorkerInfos struct {
	Infos map[string]WorkerInfo
}

func GetWorkerInfos(ch chan WorkerInfos) {
	go func() {
		info := WorkerInfos{
			Infos: map[string]WorkerInfo{},
		}

		out, err := systemapi.RunCommand(exec.Command("/usr/local/bin/lotus-miner", "sealing", "workers"))
		if err != nil {
			log.Errorf(log.Fields{}, "fail to run lotus-miner sealing workers: %v", err)
			ch <- info
			return
		}

		br := bufio.NewReader(bytes.NewReader(out))
		curWorker := ""

		for {
			line, _, err := br.ReadLine()
			if err != nil {
				break
			}

			lineStr := string(line)
			status := ""

			if strings.HasPrefix(lineStr, "Worker ") {
				hostStr := strings.Split(lineStr, ", host ")[1]
				hostStrs := strings.Split(hostStr, "/")
				if len(hostStrs) < 2 {
					curWorker = hostStrs[0]
				} else {
					hostStrs = strings.Split(hostStrs[1], " ")
					curWorker = hostStrs[0]
					status = strings.Replace(hostStrs[1], "(", "", -1)
					status = strings.Replace(status, "(", "", -1)
				}
			}

			if _, ok := info.Infos[curWorker]; !ok && curWorker != "localhost" {
				maintaining := 0
				if strings.Contains(status, "M") {
					maintaining = 1
				}
				rejectTask := 0
				if strings.Contains(status, "R") {
					rejectTask = 1
				}
				info.Infos[curWorker] = WorkerInfo{
					Maintaining: maintaining,
					RejectTask:  rejectTask,
				}
			}

			workerInfo := info.Infos[curWorker]

			if strings.Contains(lineStr, "GPU: ") && curWorker != "localhost" {
				workerInfo.GPUs += 1
			}
			if curWorker != "localhost" {
				info.Infos[curWorker] = workerInfo
			}
		}

		ch <- info
	}()
}

func convertTiB(value float64, line string) float64 {
	if strings.Contains(line, "Pi") {
		value = value * 1024
	} else if strings.Contains(line, "Gi") {
		value = value / 1024
	}
	return value
}
