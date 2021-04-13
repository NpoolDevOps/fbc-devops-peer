package devopsruntime

import (
	"bufio"
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/jaypipes/ghw"
	"github.com/rai-project/nvidia-smi"
	"os/exec"
	"strings"
)

func rootInDisk(disk *ghw.Disk) bool {
	for _, part := range disk.Partitions {
		if part.MountPoint == "/" {
			return true
		}
	}
	return false
}

func getNvmeList() []string {
	block, _ := ghw.Block()

	nvmes := []string{}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if strings.Contains(disk.Name, "nvme") {
			nvmes = append(nvmes, disk.Name)
		}
	}

	return nvmes
}

func GetNvmeCount() (int, error) {
	return len(getNvmeList()), nil
}

type diskInfo struct {
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	Sn     string `json:"sn"`
	Wwn    string `json:"wwn"`
}

func GetNvmeDesc() ([]string, error) {
	nvmeDescs := []string{}

	block, _ := ghw.Block()
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if !strings.Contains(disk.Name, "nvme") {
			continue
		}
		info := diskInfo{
			Name:   disk.Name,
			Vendor: disk.Vendor,
			Model:  disk.Model,
			Sn:     disk.SerialNumber,
			Wwn:    disk.WWN,
		}
		b, _ := json.Marshal(&info)
		nvmeDescs = append(nvmeDescs, string(b))
	}

	return nvmeDescs, nil
}

type gpuDesc struct {
	Vendor        string `json:"vendor"`
	Product       string `json:"product"`
	DriverVersion string `json:"driver_version"`
}

func GetGpuCount() (int, error) {
	gpu, _ := ghw.GPU()
	return len(gpu.GraphicsCards), nil
}

func GetGpuDesc() ([]string, error) {
	gpus := []string{}

	nvgpu, err := nvidiasmi.New()
	if err == nil {
		for _, gpu := range nvgpu.GPUS {
			info := gpuDesc{
				Vendor:        "NVIDIA",
				Product:       gpu.ProductName,
				DriverVersion: nvgpu.DriverVersion,
			}
			desc, _ := json.Marshal(info)
			gpus = append(gpus, string(desc))
		}
	}

	gpu, _ := ghw.GPU()
	for _, card := range gpu.GraphicsCards {
		added := false
		for _, name := range gpus {
			if strings.Contains(name, card.DeviceInfo.Product.Name) {
				added = true
			}
		}
		if added {
			continue
		}
		info := gpuDesc{
			Vendor:  card.DeviceInfo.Vendor.Name,
			Product: card.DeviceInfo.Product.Name,
		}
		desc, _ := json.Marshal(info)
		gpus = append(gpus, string(desc))
	}

	return gpus, nil
}

func GetMemoryCount() (int, error) {
	memory, _ := ghw.Memory()

	if memory.Modules == nil || len(memory.Modules) == 0 {
		spec := machspec.NewMachineSpec()
		err := spec.PrepareLowLevel()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to prepare spec: %v", err)
		}
		return len(spec.Memory), nil
	} else {
		return len(memory.Modules), nil
	}

	return 0, nil
}

func GetMemorySize() (uint64, error) {
	memory, _ := ghw.Memory()
	return uint64(memory.TotalPhysicalBytes), nil
}

func GetMemoryDesc() ([]string, error) {
	mems := []string{}

	memory, _ := ghw.Memory()
	for _, m := range memory.Modules {
		info := machspec.Memory{
			Dimm:         m.Location,
			Sn:           m.SerialNumber,
			SizeGB:       int(m.SizeBytes / 1024 / 1024 / 1024),
			Manufacturer: m.Vendor,
		}
		b, _ := json.Marshal(info)
		mems = append(mems, string(b))
	}

	if len(mems) == 0 {
		spec := machspec.NewMachineSpec()
		err := spec.PrepareLowLevel()
		if err == nil {
			for _, m := range spec.Memory {
				b, _ := json.Marshal(m)
				mems = append(mems, string(b))
			}
		}
	}

	return mems, nil
}

func GetCpuCount() (int, error) {
	cpu, _ := ghw.CPU()
	return len(cpu.Processors), nil
}

type cpuDesc struct {
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
}

func GetCpuDesc() ([]string, error) {
	cpu, _ := ghw.CPU()

	cpus := []string{}
	for _, c := range cpu.Processors {
		desc := cpuDesc{
			Vendor: c.Vendor,
			Model:  c.Model,
		}
		b, _ := json.Marshal(desc)
		cpus = append(cpus, string(b))
	}

	return cpus, nil
}

func getHddList() []string {
	block, _ := ghw.Block()

	hdds := []string{}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if strings.Contains(disk.Name, "nvme") {
			continue
		}
		hdds = append(hdds, disk.Name)
	}

	return hdds
}

func GetHddCount() (int, error) {
	return len(getHddList()), nil
}

func GetHddDesc() ([]string, error) {
	hddDescs := []string{}

	block, _ := ghw.Block()
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if strings.Contains(disk.Name, "nvme") {
			continue
		}
		info := diskInfo{
			Name:   disk.Name,
			Vendor: disk.Vendor,
			Model:  disk.Model,
			Sn:     disk.SerialNumber,
			Wwn:    disk.WWN,
		}
		b, _ := json.Marshal(&info)
		hddDescs = append(hddDescs, string(b))
	}

	return hddDescs, nil
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

func GetEthernetCount() (int, error) {
	out, err := exec.Command("lshw", "-C", "network").Output()
	if err != nil {
		return 0, err
	}

	br := bufio.NewReader(strings.NewReader(string(out)))
	count := 0
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "description:") {
			count += 1
		}
	}
	return count, nil
}

func GetEthernetDesc() ([]string, error) {
	out, err := exec.Command("lshw", "-C", "network").Output()
	if err != nil {
		return nil, err
	}

	eths := []string{}
	br := bufio.NewReader(strings.NewReader(string(out)))
	eth := Ethernet{}
	parsed := false
	hasNetwork := false

	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}

		if strings.Contains(string(line), "*-network") && parsed {
			b, _ := json.Marshal(eth)
			eths = append(eths, string(b))
			eth = Ethernet{}
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
		b, _ := json.Marshal(eth)
		eths = append(eths, string(b))
	}

	return eths, nil
}
