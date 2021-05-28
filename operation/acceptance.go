package operation

import (
	_ "encoding/json"
)

type bondConfig struct {
	Ethernets []string `json:"ethernets"`
	Bond      string   `json:"bond"`
}

type acceptanceParams struct {
	Cpus               int        `json:"cpus"`
	CpuBrand           string     `json:"cpu_brand"`
	CpuCores           int        `json:"cpu_cores"`
	Memorys            int        `json:"memorys"`
	MemoryUnitSize     string     `json:"memory_unit_size"`
	Gpus               int        `json:"gpus"`
	GpuBrand           string     `json:"gpu_brand"`
	Nvmes              int        `json:"nvmes"`
	NvmeUnitSize       string     `json:"nvme_unit_size"`
	Hdds               int        `json:"hdds"`
	HddUnitSize        string     `json:"hdd_unit_size"`
	Ethernets          int        `json:"ethernets"`
	EthernetSpeed      string     `json:"ethernet_speed"`
	EthernetBondConfig bondConfig `json:"ethernet_bond_config"`
}

func acceptanceExec(params string) (interface{}, error) {
	return nil, nil
}
