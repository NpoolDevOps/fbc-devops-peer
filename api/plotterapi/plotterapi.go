package plotterapi

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/minerapi"
)

func GetPlotterProcessNum() (int64, error) {
	out, err := minerapi.RunCommand(exec.Command("ps", "-ef"))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get process...%v", err)
		return 0, err
	}

	br := bufio.NewReader(bytes.NewReader(out))
	count := 0
	str := ""
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		linestr := strings.TrimSpace(string(line))
		if strings.Contains(linestr, "/usr/local/bin/ProofOfSpace") {
			lineBefore := strings.Split(linestr, "-i")[1]
			trueLine := strings.TrimSpace(strings.Split(lineBefore, "-m")[0])
			str = str + trueLine
			if strings.Count(str, trueLine) == 1 {
				count++
			}
		}
	}
	return int64(count), nil
}

func GetStatus(role string) (string, error) {
	out, err := minerapi.RunCommand(exec.Command("ps", "-ef"))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get process...%v", err)
		return "", err
	}

	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		linestr := strings.TrimSpace(string(line))
		if strings.Contains(linestr, role) {
			return "active", nil
		}
	}
	return "dead", nil

}

func GetPlotterTime() (float64, int64, error) {
	out, err := minerapi.RunCommand(exec.Command("grep", "Use", "/var/log/chia-plotter.log"))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get plotter time...%v", err)
		return 0, 0, err
	}
	var totalTime float64 = 0
	var count int64 = 0

	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "msg=\"Use") && !strings.Contains(string(line), "Failed to connect") {
			time := strings.TrimSpace(strings.Split(string(line), "Use")[1])
			timeArr := strings.Split(time, "m")
			timeMin, _ := strconv.ParseFloat(timeArr[0], 64)
			timeSec, _ := strconv.ParseFloat(strings.TrimSpace(strings.Split(timeArr[1], "s")[0]), 64)
			totalTime = totalTime + timeMin*60 + timeSec
			count++
		}
	}

	avgTime := totalTime / float64(count)
	return avgTime, count, nil
}
