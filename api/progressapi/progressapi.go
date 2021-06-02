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
	lines = lines -1
	return lines, nil
}

func GetProgressInfo(device string) int64 {
	pid, err := getProgressPid(device)
	if err != nil {
		log.Errorf(log.Fields{}, "fail, error is: %v", err)
	}
	fileOpened, err := getProgressFileOpened(pid)
	if err != nil {
		log.Errorf(log.Fields{}, "fail, error is: %v", err)
	}
	log.Infof(log.Fields{}, "%v file open number is: %v", device, fileOpened)
	return fileOpened
}
