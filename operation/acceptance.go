package operation

import (
	"encoding/json"
	"fmt"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	"github.com/docker/go-units"
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

func kernelError(specs []string) []acceptanceResult {
	results := []acceptanceResult{}

	parser, err := kmsgparser.NewParser()
	if err != nil {
		results = append(results, newAcceptanceResult("Kernel Error", err.Error(), "", err))
		return results
	}

	msgCh := parser.Parse()
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
			for _, spec := range specs {
				_, ok := specMap[spec]
				if ok && strings.Contains(msg.Message, spec) {
					specMap[spec] = struct{}{}
					results = append(results, newAcceptanceResult("Kernel Error", "", msg.Message, err))
				}
			}
		}
	}

	return results
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

		cpuList := runtime.GetCpuList()
		for i, cpu := range cpuList {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("CPU %v Desc", i), p.CpuBrand, cpu.Model, err))
		}
	}

	memoryErr := kernelError([]string{
		"CE memory read error",
	})
	results.Results = append(results.Results, memoryErr[0:]...)

	if 0 < p.Nvmes {
		nvmes, err := runtime.GetNvmeCount()
		results.Results = append(results.Results, newAcceptanceResult("NVME Count", p.Nvmes, nvmes, err))

		nvmeList := runtime.GetNvmeList()

		nvmeUnitBytes, err := units.RAMInBytes(p.NvmeUnitSize)
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("NVME Unit Bytes", p.NvmeUnitSize, "0", err))
		} else {
			nvmeUnitBytes = nvmeUnitBytes / 1024 / 1024 / 1024 / 1024

			for i, nvme := range nvmeList {
				sizeBytes := uint64(nvme.SizeBytes) / 1024 / 1024 / 1024 / 1024
				results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("NVME %v Desc %v", i, p.NvmeUnitSize), nvmeUnitBytes, sizeBytes, err))
			}
		}
	}

	if 0 < p.Hdds {
		hdds, err := runtime.GetHddCount()
		results.Results = append(results.Results, newAcceptanceResult("HDD Count", p.Hdds, hdds, err))

		hddList := runtime.GetHddList()

		hddUnitBytes, err := units.RAMInBytes(p.HddUnitSize)
		if err != nil {
			results.Results = append(results.Results, newAcceptanceResult("HDD Unit Bytes", p.HddUnitSize, "0", err))
		} else {
			hddUnitBytes1 := uint64(hddUnitBytes) / 1024 / 1024 / 1024 / 1024

			for i, hdd := range hddList {
				sizeBytes := uint64(hdd.SizeBytes) / 1024 / 1024 / 1024 / 1024
				results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("HDD %v Desc %v", i, p.HddUnitSize), hddUnitBytes1, sizeBytes, err))
			}
		}
	}

	if 0 < p.Gpus {
		gpus, err := runtime.GetGpuCount()
		results.Results = append(results.Results, newAcceptanceResult("GPU Count", p.Gpus, gpus, err))

		gpuList := runtime.GetGpuList()
		for i, gpu := range gpuList {
			results.Results = append(results.Results, newAcceptanceResult(fmt.Sprintf("GPU %v Desc", i), p.GpuBrand, gpu.DeviceInfo.Product.Name, err))
		}
	}

	if 0 < p.Ethernets {
		eths, err := runtime.GetEthernetCount()
		results.Results = append(results.Results, newAcceptanceResult("Ethernet Count", p.Ethernets, eths, err))
	}

	// If no memory error, do simple NVME | HDD test to check IO error
	// If nvme or hdd is mounted, notify to deployer to check, or pass force to umount and test them
	// Simple test nvme and collect test result, and kernel error
	// Simple test hdd and collect test result, and kernel error

	// Check Ethernet

	return results, nil
}
