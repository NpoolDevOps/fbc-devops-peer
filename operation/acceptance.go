package operation

import (
	"encoding/json"
	"fmt"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	"github.com/euank/go-kmsg-parser/kmsgparser"
	"strings"
	"time"
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
	OsSpec             string     `json:"os_spec"`
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
		if 0 < len(expect.(string)) {
			testOk = strings.Contains(result.(string), expect.(string))
		}
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

	if 0 < p.Cpus {
		cpus, err := runtime.GetCpuCount()
		results.Results = append(results.Results, newAcceptanceResult("CPU Count", p.Cpus, cpus, err))

		cpuDesc, err := runtime.GetCpuDesc()
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("CPU Desc", p.CpuBrand, "", err))
		}

		for i, desc := range cpuDesc {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("CPU %v Desc", i), p.CpuBrand, desc, err))
		}
	}

	parser, err := kmsgparser.NewParser()
	if err != nil {
		results.Results = append(results.Results, newAcceptanceResult("Kernel Error", err.Error(), "", err))
	}

	msgCh := parser.Parse()
	errSpec := []string{
		"CE memory read error",
	}
	specMap := map[string]struct{}{}

	go func() {
		time.Sleep(10 * time.Second)
		parser.Close()
	}()

processDmesgLoop:
	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				break processDmesgLoop
			}
			for _, spec := range errSpec {
				_, ok := specMap[spec]
				if ok && strings.Contains(msg.Message, spec) {
					specMap[spec] = struct{}{}
					results.Results = append(results.Results, newAcceptanceResult("Kernel Error", "", msg.Message, err))
				}
			}
		}
	}

	// If no memory error, do simple NVME | HDD test to check IO error
	// If nvme or hdd is mounted, notify to deployer to check, or pass force to umount and test them
	// Get nvme count
	// Get nvme parameter
	// Simple test nvme and collect test result, and kernel error
	// Get hdd count
	// Get hdd parameter
	// Simple test hdd and collect test result, and kernel error

	// Check GPU
	// Check Ethernet

	return results, nil
}
