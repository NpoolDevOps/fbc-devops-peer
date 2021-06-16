package systemapi

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	"github.com/moby/sys/mountinfo"
	"golang.org/x/xerrors"
)

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func FilePerm2Int(file string) (int64, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	strMode := fmt.Sprintf("%o", fi.Mode().Perm())
	mode, _ := strconv.ParseInt(strMode, 10, 64)
	return mode, nil
}

func MountpointWrittable(mountpoint string) (bool, error) {
	var info *mountinfo.Info = nil

	_, err := mountinfo.GetMounts(func(i *mountinfo.Info) (skip, stop bool) {
		if info.Mountpoint == mountpoint {
			info = i
			return false, true
		}
		return false, false
	})

	if err != nil {
		return false, err
	}

	if info != nil {
		if strings.TrimSpace(strings.Split(info.Options, ",")[0]) == "rw" {
			return true, nil
		}
		return false, nil
	}

	return false, xerrors.Errorf("mountpoint %v not found", mountpoint)
}

func StatSubDirs(dir string, sublevel int) map[string]error {
	stat := map[string]error{}
	mySlashes := strings.Count(dir, "/")

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if strings.Count(path, "/") == sublevel+mySlashes {
			stat[path] = err
		}
		return nil
	})

	return stat
}

func GetNvmeTemperature(nvme string) (map[string]float64, error) {
	temperatureList := make(map[string]float64)
	out, err := RunCommand(exec.Command("nvme", "smart-log", nvme))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to run nvme info, %v", err)
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if !strings.Contains(string(line), " Temperature ") {
			if strings.Contains(string(line), "temperature") || strings.Contains(string(line), "Temperature Sensor") {
				temperatureName := strings.TrimSpace(strings.Split(string(line), ":")[0])
				temperatureBefore := strings.TrimSpace(strings.Split(string(line), ":")[1])
				temperature := strings.TrimSpace(strings.Split(temperatureBefore, " ")[0])
				temperature2Float, _ := strconv.ParseFloat(temperature, 64)
				temperatureList[temperatureName] = temperature2Float
			}
		}
	}
	return temperatureList, nil
}

func GetNvmeTemperatureList() (map[string]map[string]float64, error) {
	nvmeTemperatureList := make(map[string]map[string]float64)
	nvmeList := runtime.GetNvmeList()
	for _, nvme := range nvmeList {
		temperatureList := make(map[string]float64)
		var err error
		temperatureList, err = GetNvmeTemperature("/dev/" + nvme.Name)
		if err != nil {
			return nil, err
		}
		nvmeTemperatureList[nvme.Name] = temperatureList
	}
	return nvmeTemperatureList, nil
}
