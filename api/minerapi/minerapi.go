package minerapi

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type MinerInfo struct {
	MinerId    string
	SectorSize string
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

		out, _ := exec.Command("lotus-miner", "info", hideSector).Output()
		br := bufio.NewReader(bytes.NewReader(out))

		info := MinerInfo{
			State: map[string]uint64{},
		}

		for {
			line, _, err := br.ReadLine()
			if err != nil {
				break
			}

			lineStr := strings.TrimSpace(string(line))

			if strings.Contains(lineStr, "Miner: ") {
				info.MinerId = strings.Split(lineStr, " ")[1]
				info.SectorSize = strings.Split(lineStr, "(")[1]
				info.SectorSize = strings.Split(info.SectorSize, ")")[0]
			}
			if strings.Contains(lineStr, "Power: ") {
				info.Power, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
			}
			if strings.Contains(lineStr, "Raw: ") {
				info.Raw, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
			}
			if strings.Contains(lineStr, "Committed: ") {
				info.Committed, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
			}
			if !inSectorState {
				if strings.Contains(lineStr, "Proving: ") {
					info.Proving, _ = strconv.ParseFloat(strings.Split(lineStr, " ")[1], 64)
					if strings.Contains(lineStr, "Faulty, ") {
						faulty := strings.Split(lineStr, "(")[1]
						info.Faulty, _ = strconv.ParseFloat(strings.Split(faulty, " ")[0], 64)
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
