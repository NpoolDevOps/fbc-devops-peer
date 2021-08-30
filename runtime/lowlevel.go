package devopsruntime

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"strings"

	machspec "github.com/EntropyPool/machine-spec"
	"github.com/jaypipes/ghw"
	block2 "github.com/jaypipes/ghw/pkg/block"
	cpu2 "github.com/jaypipes/ghw/pkg/cpu"
)

func rootInDisk(disk *ghw.Disk) bool {
	for _, part := range disk.Partitions {
		if part.MountPoint == "/" {
			return true
		}
	}
	return false
}

type DiskInfo = ghw.Disk

func Info2String(i interface{}) string {
	info, _ := json.Marshal(i)
	return string(info)
}

func GetNvmeList() []*DiskInfo {
	block, err := ghw.Block()
	if err != nil {
		return nil
	}

	disks := []*DiskInfo{}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if disk.StorageController.String() == block2.STORAGE_CONTROLLER_NVME.String() {
			disks = append(disks, disk)
		}
	}

	return disks
}

func GetHddList() []*DiskInfo {
	block, err := ghw.Block()
	if err != nil {
		return nil
	}

	disks := []*DiskInfo{}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if disk.DriveType.String() == block2.DRIVE_TYPE_HDD.String() {
			disks = append(disks, disk)
		}
	}

	return disks
}

func GetSsdList() []*DiskInfo {
	block, err := ghw.Block()
	if err != nil {
		return nil
	}
	disks := []*DiskInfo{}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if disk.DriveType.String() == block2.DRIVE_TYPE_SSD.String() {
			disks = append(disks, disk)
		}
	}
	return disks
}

type GpuInfo = ghw.GraphicsCard

func GetGpuList() []*GpuInfo {
	gpu, _ := ghw.GPU()
	return gpu.GraphicsCards
}

func GetMemoryList() []machspec.Memory {
	mems := []machspec.Memory{}

	memory, _ := ghw.Memory()
	for _, m := range memory.Modules {
		info := machspec.Memory{
			Dimm:         m.Location,
			Sn:           m.SerialNumber,
			SizeGB:       int(m.SizeBytes / 1024 / 1024 / 1024),
			Manufacturer: m.Vendor,
		}
		mems = append(mems, info)
	}

	if len(mems) == 0 {
		spec := machspec.NewMachineSpec()
		if spec == nil {
			return []machspec.Memory{}
		}

		err := spec.PrepareLowLevel()
		if err == nil {
			return spec.Memory
		}
	}

	return mems
}

type CpuInfo = cpu2.Processor

func GetCpuList() []*CpuInfo {
	cpu, _ := ghw.CPU()
	return cpu.Processors
}

type EthernetInfo struct {
	Description   string `json:"description"`
	Vendor        string `json:"vendor"`
	LogicName     string `json:"name"`
	Serial        string `json:"serial"`
	Configuration string `json:"configuration"`
	Ip            string `json:"ip"`
	BusInfo       string `json:"bus_info"`
	Capacity      string `json:"capacity"`
	IsExporter    bool   `json:"is_exporter"`
}

func GetEthernetList() []*EthernetInfo {
	out, err := exec.Command("lshw", "-C", "network").Output()
	if err != nil {
		return nil
	}

	eths := []*EthernetInfo{}

	br := bufio.NewReader(strings.NewReader(string(out)))
	eth := &EthernetInfo{}
	parsed := false
	hasNetwork := false

	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}

		if strings.Contains(string(line), "*-network") && parsed {
			eths = append(eths, eth)
			eth = &EthernetInfo{}
			hasNetwork = true
		}

		if strings.Contains(string(line), "description:") {
			eth.Description = strings.Split(string(line), ": ")[1]
		}
		if strings.Contains(string(line), "vendor:") {
			eth.Vendor = strings.Split(string(line), ": ")[1]
		}
		if strings.Contains(string(line), "logic name:") {
			eth.LogicName = strings.Split(string(line), ": ")[1]
		}
		if strings.Contains(string(line), "serial:") {
			eth.Serial = strings.Split(string(line), ": ")[1]
		}
		if strings.Contains(string(line), "bus info:") {
			eth.BusInfo = strings.Split(string(line), ": ")[1]
		}
		if strings.Contains(string(line), "configuration:") {
			eth.Configuration = strings.Split(string(line), ": ")[1]

			ips := strings.Split(eth.Configuration, "ip=")
			if 1 < len(ips) {
				eth.Ip = strings.Split(ips[1], " ")[0]
			}
		}
		if strings.Contains(string(line), "capacity:") {
			eth.Capacity = strings.Split(string(line), ": ")[1]
		}
		if strings.Contains(string(line), "size:") {
			eth.Capacity = strings.Split(string(line), ": ")[1]
		}

		parsed = true
	}

	if hasNetwork {
		eths = append(eths, eth)
	}

	return eths
}
