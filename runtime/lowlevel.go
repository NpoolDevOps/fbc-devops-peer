package runtime

import (
	"bufio"
	"os/exec"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/jaypipes/ghw"
	nvidiasmi "github.com/rai-project/nvidia-smi"
)

type DiskInfo struct {
	Name       string           `json:"name"`
	Vendor     string           `json:"vendor"`
	Model      string           `json:"model"`
	Sn         string           `json:"sn"`
	Wwn        string           `json:"wwn"`
	Size       uint64           `json:"size"`
	Partitions []*ghw.Partition `json:"partitions"`
}

type CpuDesc struct {
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
}

type CpuInfo struct {
	Count   int       `json:"count"`
	CpuDesc []CpuDesc `json:"cpu_descs"`
}

type GpuInfo struct {
	Count   int       `json:"count"`
	GpuDesc []GpuDesc `json:"gpu_desc"`
}
type GpuDesc struct {
	Vendor        string `json:"vendor"`
	Product       string `json:"product"`
	DriverVersion string `json:"driver_version"`
}

type MemoryInfo struct {
	Count      int               `json:"count"`
	Size       uint64            `json:"size"`
	MemoryDesc []machspec.Memory `json:"memory_desc"`
}

type Ethernet struct {
	Description   string `json:"description"`
	Vendor        string `json:"vendor"`
	LogicName     string `json:"name"`
	Serial        string `json:"serial"`
	Configuration string `json:"configuration"`
	Ip            string `json:"ip"`
	BusInfo       string `json:"bus_info"`
	Capacity      string `json:"capacity"`
}

type EthernetInfo struct {
	Count        int        `json:"count"`
	Speed        string     `json:"speed"`
	EthernetDesc []Ethernet `json:"ethernet_desc"`
}

func rootInDisk(disk *ghw.Disk) bool {
	for _, part := range disk.Partitions {
		if part.MountPoint == "/" {
			return true
		}
	}
	return false
}

func GetHdds() ([]DiskInfo, error) {
	hddsInfo := []DiskInfo{}

	block, err := ghw.Block()
	if err != nil {
		return nil, err
	}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if disk.DriveType.String() != "HDD" {
			continue
		}
		info := DiskInfo{}
		if disk.DriveType.String() == "HDD" {
			info.Name = disk.Name
			info.Vendor = disk.Vendor
			info.Model = disk.Model
			info.Sn = disk.SerialNumber
			info.Wwn = disk.WWN
			info.Size = disk.SizeBytes
			info.Partitions = disk.Partitions
		}
		hddsInfo = append(hddsInfo, info)

	}
	return hddsInfo, nil
}

func GetNvmes() ([]DiskInfo, error) {
	nvmesInfo := []DiskInfo{}

	block, _ := ghw.Block()
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if disk.StorageController.String() != "NVMe" {
			continue
		}
		info := DiskInfo{}
		if disk.StorageController.String() == "NVMe" {
			info.Name = disk.Name
			info.Vendor = disk.Vendor
			info.Model = disk.Model
			info.Sn = disk.SerialNumber
			info.Wwn = disk.WWN
			info.Size = disk.SizeBytes
			info.Partitions = disk.Partitions
		}
		nvmesInfo = append(nvmesInfo, info)

	}
	return nvmesInfo, nil
}

func GetSsds() ([]DiskInfo, error) {
	ssdsInfo := []DiskInfo{}

	block, _ := ghw.Block()
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if disk.DriveType.String() != "SSD" {
			continue
		}
		if disk.DriveType.String() == "SSD" {
			info := DiskInfo{
				Name:       disk.Name,
				Vendor:     disk.Vendor,
				Model:      disk.Model,
				Sn:         disk.SerialNumber,
				Wwn:        disk.WWN,
				Size:       disk.SizeBytes,
				Partitions: disk.Partitions,
			}
			ssdsInfo = append(ssdsInfo, info)
		}

	}
	return ssdsInfo, nil
}

func GetMemorys() (MemoryInfo, error) {
	mems := MemoryInfo{}
	memory, _ := ghw.Memory()
	mems.Size = uint64(memory.TotalPhysicalBytes)

	if memory.Modules == nil || len(memory.Modules) == 0 {
		spec := machspec.NewMachineSpec()
		err := spec.PrepareLowLevel()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to prepare spec: %v", err)
		}
		mems.Count = len(spec.Memory)
	} else {
		mems.Count = len(memory.Modules)
	}

	for _, m := range memory.Modules {
		info := machspec.Memory{
			Dimm:         m.Location,
			Sn:           m.SerialNumber,
			SizeGB:       int(m.SizeBytes / 1024 / 1024 / 1024),
			Manufacturer: m.Vendor,
		}
		mems.MemoryDesc = append(mems.MemoryDesc, info)
	}

	if len(mems.MemoryDesc) == 0 {
		spec := machspec.NewMachineSpec()
		err := spec.PrepareLowLevel()
		if err == nil {
			mems.MemoryDesc = append(mems.MemoryDesc, spec.Memory...)
		}
	}

	return mems, nil
}

func GetCpus() (CpuInfo, error) {
	cpus := CpuInfo{}
	cpu, _ := ghw.CPU()
	cpus.Count = len(cpu.Processors)
	for _, c := range cpu.Processors {
		desc := CpuDesc{
			Vendor: c.Vendor,
			Model:  c.Model,
		}
		cpus.CpuDesc = append(cpus.CpuDesc, desc)
	}
	return cpus, nil
}

func GetGpus() (GpuInfo, error) {
	gpus := GpuInfo{}
	nvgpu, err := nvidiasmi.New()
	if err == nil {
		for _, gpu := range nvgpu.GPUS {
			desc := GpuDesc{
				Vendor:        "NVIDIA",
				Product:       gpu.ProductName,
				DriverVersion: nvgpu.DriverVersion,
			}
			gpus.GpuDesc = append(gpus.GpuDesc, desc)
		}
	}

	gpu, _ := ghw.GPU()
	gpus.Count = len(gpu.GraphicsCards)
	for _, card := range gpu.GraphicsCards {
		added := false
		for _, g := range gpus.GpuDesc {
			if g.Product == card.DeviceInfo.Product.Name {
				added = true
			}
		}
		if added {
			continue
		}
		desc := GpuDesc{
			Vendor:  card.DeviceInfo.Vendor.Name,
			Product: card.DeviceInfo.Product.Name,
		}
		gpus.GpuDesc = append(gpus.GpuDesc, desc)
	}
	return gpus, nil
}

func GetEthernet() (EthernetInfo, error) {
	out, _ := exec.Command("lshw", "-C", "network").Output()
	br := bufio.NewReader(strings.NewReader(string(out)))
	count := 0
	ethernets := EthernetInfo{}
	eth := Ethernet{}
	parsed := false
	hasNetwork := false
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "*-network") && parsed {
			ethernets.EthernetDesc = append(ethernets.EthernetDesc, eth)
			eth = Ethernet{}
			hasNetwork = true
		}
		if strings.Contains(string(line), "description:") {
			eth.Description = strings.Split(string(line), ": ")[1]
			count += 1
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
			if strings.Contains(eth.Configuration, "speed") {
				ethArr := strings.Split(eth.Configuration, "speed=")
				ethernets.Speed = strings.TrimSpace(ethArr[1])
			}

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
		ethernets.EthernetDesc = append(ethernets.EthernetDesc, eth)
	}
	ethernets.Count = count
	return ethernets, nil
}
