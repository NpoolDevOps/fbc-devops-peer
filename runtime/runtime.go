package runtime

import (
	"encoding/json"
)

//NVMe
func GetNvmeList() []string {
	nvmes := []string{}
	nvmesList, _ := GetNvmes()
	for _, nvme := range nvmesList {
		nvmes = append(nvmes, nvme.Name)
	}
	return nvmes
}

func GetNvmeDesc() ([]string, error) {
	nvmeDescs := []string{}
	nvmeList, _ := GetNvmes()
	for _, nvme := range nvmeList {
		b, _ := json.Marshal(&nvme)
		nvmeDescs = append(nvmeDescs, string(b))
	}
	return nvmeDescs, nil
}

func GetNvmeCount() (int, error) {
	return len(GetNvmeList()), nil
}

//GPU
func GetGpuCount() (int, error) {
	gpus, _ := GetGpus()
	return gpus.Count, nil
}

func GetGpuDesc() ([]string, error) {
	gpus := []string{}
	gpu, _ := GetGpus()
	for _, g := range gpu.GpuDesc {
		b, _ := json.Marshal(g)
		gpus = append(gpus, string(b))
	}
	return gpus, nil
}

//memory
func GetMemoryCount() (int, error) {
	mems, _ := GetMemorys()
	return mems.Count, nil
}

func GetMemorySize() (uint64, error) {
	mems, _ := GetMemorys()
	return mems.Size, nil
}

func GetMemoryDesc() ([]string, error) {
	mems := []string{}
	memory, _ := GetMemorys()
	for _, m := range memory.MemoryDesc {
		b, _ := json.Marshal(m)
		mems = append(mems, string(b))
	}
	return mems, nil
}

//cpu
func GetCpuCount() (int, error) {
	cpus, _ := GetCpus()
	return cpus.Count, nil
}

func GetCpuDesc() ([]string, error) {
	cpus := []string{}
	cpu, _ := GetCpus()
	for _, c := range cpu.CpuDesc {
		b, _ := json.Marshal(c)
		cpus = append(cpus, string(b))
	}
	return cpus, nil
}

//hdd
func GetHddList() []string {
	hdds := []string{}
	hddsList, _ := GetHdds()
	for _, hdd := range hddsList {
		hdds = append(hdds, hdd.Name)
	}
	return hdds
}

func GetHddCount() (int, error) {
	return len(GetHddList()), nil
}

func GetHddDesc() ([]string, error) {
	hddDescs := []string{}
	hddsList, err := GetHdds()
	if err != nil {
		return nil, err
	}
	for _, hdd := range hddsList {
		b, err := json.Marshal(&hdd)
		if err != nil {
			hddDescs = append(hddDescs, err.Error())
			return hddDescs, err
		}
		hddDescs = append(hddDescs, string(b))
	}
	return hddDescs, nil
}

//SSD
func GetSsdsList() []string {
	ssds := []string{}
	ssdsList, _ := GetSsds()
	for _, ssd := range ssdsList {
		ssds = append(ssds, ssd.Name)
	}
	return ssds
}

func GetSsdsCount() (int, error) {
	return len(GetSsdsList()), nil
}

func GetSsdsDesc() ([]string, error) {
	ssdsDescs := []string{}
	ssdsList, _ := GetSsds()
	for _, ssd := range ssdsList {
		b, _ := json.Marshal(&ssd)
		ssdsDescs = append(ssdsDescs, string(b))
	}
	return ssdsDescs, nil
}

//Ethernet
func GetEthernetCount() (int, error) {
	ethernet, _ := GetEthernet()
	return ethernet.Count, nil
}

func GetEthernetDesc() ([]string, error) {
	eths := []string{}
	ethernet, _ := GetEthernet()
	for _, v := range ethernet.EthernetDesc {
		b, _ := json.Marshal(v)
		eths = append(eths, string(b))
	}
	return eths, nil
}
