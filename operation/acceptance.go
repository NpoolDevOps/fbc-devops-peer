package operation

import (
	"encoding/json"
	"fmt"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
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

type acceptanceResult struct {
	Result      string `json:"result"`
	TestName    string `json:"test_name"`
	Description string `json:"description"`
}

type acceptanceResults struct {
	Results []acceptanceResult `json:"acceptance_results"`
}

func newAcceptanceIntResult(name string, expect, result int, err string) acceptanceResult {
	if expect == result {
		return acceptanceResult{
			Result:      "OK",
			TestName:    name,
			Description: fmt.Sprintf("CHECK %v [%v = %v]", name, expect, result),
		}
	}

	return acceptanceResult{
		Result:      "ERROR",
		TestName:    name,
		Description: fmt.Sprintf("CHECK %v [%v != %v](%v)", name, expect, result, err),
	}
}

func acceptanceExec(params string) (interface{}, error) {
	p := acceptanceParams{}
	err := json.Unmarshal([]byte(params), &p)
	if err != nil {
		return nil, err
	}

	results := acceptanceResults{
		Results: []acceptanceResult{},
	}

	cpus, err := runtime.GetCpuCount()
	if err != nil {
		results.Results = append(results.Results, newAcceptanceIntResult("CPU Count", p.Cpus, -1, err.Error()))
	}

	if p.Cpus != cpus {
		results.Results = append(results.Results, newAcceptanceIntResult("CPU Count", p.Cpus, cpus, ""))
	}

	return nil, nil
}
