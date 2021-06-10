package operation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	"github.com/euank/go-kmsg-parser/kmsgparser"
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
	NvmeBrand          string     `json:"nvme_brand"`
	Hdds               int        `json:"hdds"`
	HddUnitSize        string     `json:"hdd_unit_size"`
	HddBrand           string     `json:"hdd_brand"`
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
	// Simple test nvme and collect test result, and kernel error
	if 0 < p.Nvmes {
		// Get nvme count
		nvmes, err := runtime.GetNvmeCount()
		results.Results = append(results.Results, newAcceptanceResult("Nvme Count", p.Nvmes, nvmes, err))
		// Get nvme parameter
		nvmeDesc, err := runtime.GetNvmeDesc()
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("Nvme Desc", p.NvmeBrand, "", err))
		}
		for i, desc := range nvmeDesc {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("Nvme %v Desc", i), p.NvmeBrand, desc, err))
		}

		nvmeInfo, _ := runtime.GetNvmes()
		for _, info := range nvmeInfo {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("Nvme %v Size", info.Name), p.NvmeUnitSize, info.Size, nil))
		}
	}

	// Simple test hdd and collect test result, and kernel error
	if 0 < p.Hdds {
		// Get hdd count
		hdds, err := runtime.GetHddCount()
		results.Results = append(results.Results, newAcceptanceResult("Hdd Count", p.Hdds, hdds, err))
		// Get hdd parameter
		hddDesc, err := runtime.GetHddDesc()
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("Hdd Desc", p.HddBrand, "", err))
		}
		for i, desc := range hddDesc {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("Hdd %v Desc", i), p.HddBrand, desc, err))
		}

		hddInfo, _ := runtime.GetHdds()
		for _, info := range hddInfo {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("Hdd %v Size", info.Name), p.HddUnitSize, info.Size, nil))
		}
	}

	// Check GPU
	if 0 < p.Gpus {
		//Get Gpu count
		gpus, err := runtime.GetGpuCount()
		results.Results = append(results.Results, newAcceptanceResult("Gpu Count", p.Gpus, gpus, err))

		//Get gpu Desc
		gpuDesc, err := runtime.GetGpuDesc()
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("Gpu Desc", p.GpuBrand, "", err))
		}
		for i, desc := range gpuDesc {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("Gpu %v Desc", i), p.GpuBrand, desc, err))
		}
	}

	// Check Ethernet
	if 0 < p.Ethernets {
		//Get Ethernet count
		ethernet, err := runtime.GetEthernetCount()
		results.Results = append(results.Results, newAcceptanceResult("Ethernet Count", p.Ethernets, ethernet, err))

		//Get ethernet speed
		ethernetSpeed, err := runtime.GetEthernetSpeed()
		results.Results = append(results.Results, newAcceptanceResult("Ethernet Speed", p.EthernetSpeed, ethernetSpeed, err))

		//Get ethernet desc
		ethernetDesc, err := runtime.GetEthernetDesc()
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("Ethernet Desc", p.EthernetBondConfig.Bond, "", err))
		}
		for i, desc := range ethernetDesc {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("Ethernet %v Desc", i), p.EthernetBondConfig.Bond, desc, err))
		}
	}

	return results, nil
}
