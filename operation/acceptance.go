package operation

import (
	"encoding/json"
	"fmt"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	"strings"
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

func newAcceptanceResult(name string, expect, result interface{}, err error) acceptanceResult {
	testOk := false

	_, ok := expect.(string)
	if ok {
		testOk = strings.Contains(result.(string), expect.(string))
	} else {
		_, ok := expect.(int)
		if ok {
			testOk = expect.(int) == result.(int)
		}
	}

	if testOk {
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
	results.Results = append(results.Results, newAcceptanceResult("CPU Count", p.Cpus, cpus, err))

	cpuDesc, err := runtime.GetCpuDesc()
	if err != nil {
		results.Results = append(results.Results, newAcceptanceResult("CPU Desc", p.CpuBrand, "", err))
	}

	for _, desc := range cpuDesc {
		results.Results = append(results.Results, newAcceptanceResult("CPU Desc", p.CpuBrand, desc, err))
	}

	return results, nil
}
