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
	"syscall"

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
		if i.Mountpoint == mountpoint {
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

func GetNvmeTemperature(nvme string) (float64, error) {
	var temperature float64 = 0
	out, err := RunCommand(exec.Command("nvme", "smart-log", nvme))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to run nvme info, %v", err)
		return 0, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if !strings.Contains(string(line), " Temperature ") {
			if strings.Contains(string(line), "temperature") || strings.Contains(string(line), "Temperature Sensor") {
				temperatureBefore := strings.TrimSpace(strings.Split(string(line), ":")[1])
				temperature2String := strings.TrimSpace(strings.Split(temperatureBefore, " ")[0])
				temperature2Float, _ := strconv.ParseFloat(temperature2String, 64)
				if temperature < temperature2Float {
					temperature = temperature2Float
				}
			}
		}
	}
	return temperature, nil
}

func GetNvmeTemperatureList() (map[string]float64, error) {
	nvmeTemperatureList := make(map[string]float64)
	nvmeList := runtime.GetNvmeList()
	for _, nvme := range nvmeList {
		var temperature float64 = 0
		var err error
		temperature, err = GetNvmeTemperature("/dev/" + nvme.Name)
		if err != nil {
			return nil, err
		}
		nvmeTemperatureList[nvme.Name] = temperature
	}
	return nvmeTemperatureList, nil
}

func GetProcessPid(process string) (string, error) {
	outPid, err := RunCommand(exec.Command("pidof", process))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v pid", process), err)
		return "", err
	}
	brPid := bufio.NewReader(bytes.NewReader(outPid))
	line, _, err := brPid.ReadLine()
	if err != nil {
		return "", err
	}
	pidArr := strings.Split(strings.TrimSpace(string(line)), " ")
	pid := strings.TrimSpace(pidArr[len(pidArr)-1])
	return pid, nil
}

func GetProcessOpenFileNumber(process string) (int64, error) {
	processPid, _ := GetProcessPid(process)
	out, err := RunCommand(exec.Command("lsof", "-p", processPid, "-n"))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v file open number", processPid), err)
		return 0, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	var openFileNumber int64 = 0
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if !strings.Contains(string(line), "USER") {
			openFileNumber += 1
		}
	}
	return openFileNumber, nil
}

func GetProcessTcpConnectNumber(process string) (int64, error) {
	out, err := RunCommand(exec.Command("netstat", "-tnlp"))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get %v TCP connect number", process), err)
		return 0, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	var tcpConnectNumber int64 = 0
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "tcp6") {
			continue
		}
		if strings.Contains(string(line), process) {
			tcpConnectNumber += 1
		}
	}
	return tcpConnectNumber, nil
}

func GetProcessCount(process string) (int64, error) {
	out, err := RunCommand(exec.Command("ps", "-ef"))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get process...%v", err)
		return 0, err
	}

	br := bufio.NewReader(bytes.NewReader(out))
	var processCount int64 = 0
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), process) {
			processCount += 1
		}
	}
	return processCount, nil
}

type DiskStatus struct {
	All  float64
	Used float64
	Free float64
}

func DiskUsage(path string) DiskStatus {
	GB := float64(1024 * 1024 * 1024)
	disk := DiskStatus{}
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		log.Errorf(log.Fields{}, "get %v status error: %v", path, err)
		return disk
	}
	disk.All = float64(fs.Blocks*uint64(fs.Bsize)) / GB
	disk.Free = float64(fs.Bfree*uint64(fs.Bsize)) / GB
	disk.Used = float64(disk.All - disk.Free)
	return disk
}

type GigabitIp struct {
	Ip string
}

type TenGigabitIp struct {
	Ip string
}

type DeviceIp struct {
	GigabitIp    GigabitIp
	TenGigabitIp TenGigabitIp
}

func GetDeviceIps() DeviceIp {
	deviceIp := DeviceIp{}
	eths := runtime.GetEthernetList()
	for _, eth := range eths {
		var capacityToFloat float64 = 0
		if eth.Capacity != "" {
			capacity := strings.TrimSpace(strings.Split(eth.Capacity, "Gbit/s")[0])
			capacityToFloat, _ = strconv.ParseFloat(capacity, 64)
		}
		if strings.Contains(eth.LogicName, "bond") || capacityToFloat >= 10 {
			deviceIp.TenGigabitIp.Ip = eth.Ip
		} else if eth.Capacity == "1Gbit/s" {
			deviceIp.GigabitIp.Ip = eth.Ip
		}
	}
	return deviceIp
}
