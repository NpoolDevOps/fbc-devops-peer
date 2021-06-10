package chiaapi

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/minerapi"
)

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
