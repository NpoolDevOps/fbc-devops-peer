package systemapi

import (
	"os"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}
