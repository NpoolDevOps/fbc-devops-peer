package systemapi

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/moby/sys/mountinfo"
)

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetFileUsageAccess(file string) float64 {
	fi, err := os.Stat(file)
	if err != nil {
		log.Errorf(log.Fields{}, "err is:", err)
		return 0
	}
	strMode := fmt.Sprintf("%o", fi.Mode().Perm())
	floatMode, _ := strconv.ParseFloat(strMode, 64)
	return floatMode
}

func GetFileMountAccess(file string) bool {
	info, _ := mountinfo.GetMounts(func(*mountinfo.Info) (skip, stop bool) {
		return false, false
	})
	var access bool
	for _, i := range info {
		if i.Mountpoint == file {
			if strings.TrimSpace(strings.Split(i.Options, ",")[0]) == "rw" {
				access = true
				break
			} else {
				access = false
				break
			}
		}
	}
	return access
}
