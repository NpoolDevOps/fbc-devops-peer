package progressapi

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	api "github.com/NpoolDevOps/fbc-devops-peer/api/minerapi"
)

func getProgressPid(name string) (string, error) {
	outPid, err := api.RunCommand(exec.Command("pidof", name))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v pid", name), err)
		return "", err
	}
	brPid := bufio.NewReader(bytes.NewReader(outPid))
	line, _, err := brPid.ReadLine()
	if err != nil {
		return "", err
	}
	linestr := strings.TrimSpace(string(line))
	return linestr, nil
}

func getProgressFileOpened(pid string) (int64, error) {
	outNum, err := api.RunCommand(exec.Command("lsof", "-p", pid, "-n"))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v file open number", pid), err)
		return 0, err
	}
	brNum := bufio.NewReader(bytes.NewReader(outNum))
	var lines int64 = 0
	for {
		_, _, err := brNum.ReadLine()
		if err != nil {
			break
		}
		lines += 1
	}
	lines = lines - 1
	return lines, nil
}

func GetProgressTcpConnects(device string) int64 {
	outTcp, err := api.RunCommand(exec.Command("netstat", "-tunlp"))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v TCP connect number", device), err)
		return 0
	}
	brTcp := bufio.NewReader(bytes.NewReader(outTcp))
	var lines int64 = 0
	for {
		line, _, err := brTcp.ReadLine()
		if err != nil {
			log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v TCP connect number", device), err)
			break
		}
		if strings.Contains(string(line), "tcp ") {
			lines += 1
		}
	}
	return lines
}

func GetProgressInfo(device string) int64 {
	pid, err := getProgressPid(device)
	if err != nil {
		log.Errorf(log.Fields{}, "fail, error is: %v", err)
		return 0
	}
	fileOpened, err := getProgressFileOpened(pid)
	if err != nil {
		log.Errorf(log.Fields{}, "fail, error is: %v", err)
		return 0
	}
	return fileOpened
}
