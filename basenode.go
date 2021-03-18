package main

import (
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	runtime "github.com/NpoolDevOps/fbc-devops-client/runtime"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
)

type Basenode struct {
	DevopsClient *DevopsClient
	PeerDesc     *PeerDesc
}

const (
	FullNode      = "fullnode"
	MinerNode     = "miner"
	FullMinerNode = "fullminer"
	WorkerNode    = "worker"
	StorageNode   = "storage"
)

type PeerHardware struct {
	NvmeCount   int    `json:"nvme_count"`
	NvmeDesc    string `json:"nvme_desc"`
	GpuCount    int    `json:"gpu_count"`
	GpuDesc     string `json:"gpu_desc"`
	MemoryCount int    `json:"memory_count"`
	MemorySize  uint64 `json:"memory_size"`
	MemoryDesc  string `json:"memory_desc"`
	CpuCount    int    `json:"cpu_count"`
	CpuDesc     string `json:"cpu_desc"`
	HddCount    int    `json:"hdd_count"`
	HddDesc     string `json:"hdd_count"`
}

type PeerDesc struct {
	MySpec       string        `json:"my_spec"`
	HardwareInfo *PeerHardware `json:"hardware_info"`
	PeerConfig   *PeerConfig   `json:"peer_config"`
}

type PeerConfig struct {
	MainRole       string        `json:"main_role"`
	SubRole        string        `json:"sub_role"`
	ParentSpec     string        `json:"parent_spec"`
	HardwareConfig *PeerHardware `json:"hardware_config"`
}

type BasenodeConfig struct {
	PeerReportAPI string
	PeerConfig    *PeerConfig
	Owner         string
}

func NewBasenode(config *BasenodeConfig) *Basenode {
	basenode := &Basenode{
		PeerDesc: &PeerDesc{
			PeerConfig: config.PeerConfig,
		},
	}

	basenode.DevopsClient = NewDevopsClient(&DevopsConfig{
		PeerReportAPI: config.PeerReportAPI,
	})
	if basenode.DevopsClient == nil {
		log.Errorf(log.Fields{}, "fail to create devops client")
		return nil
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	basenode.PeerDesc.MySpec = spec.SN()

	basenode.PeerDesc.HardwareInfo = &PeerHardware{}
	basenode.PeerDesc.HardwareInfo.UpdatePeerInfo()

	basenode.DevopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput())

	return basenode
}

func (n *Basenode) ToDeviceRegisterInput() *DeviceRegisterInput {
	return &types.DeviceRegisterInput{
		Spec:        basenode.PeerDesc.MySpec,
		ParentSpec:  basenode.PeerDesc.PeerConfig.ParentSpec,
		Role:        basenode.PeerDesc.PeerConfig.MainRole,
		SubRole:     basenode.PeerDesc.PeerConfig.SubRole,
		Owner:       basenode.Owner,
		NvmeCount:   basenode.PeerDesc.HardwareInfo.NvmeCount,
		NvmeDesc:    basenode.PeerDesc.HardwareInfo.NvmeDesc,
		GpuCount:    basenode.PeerDesc.HardwareInfo.GpuCount,
		GpuDesc:     basenode.PeerDesc.HardwareInfo.GpuDesc,
		MemoryCount: basenode.PeerDesc.HardwareInfo.MemoryCount,
		MemorySize:  basenode.PeerDesc.HardwareInfo.MemorySize,
		MemoryDesc:  basenode.PeerDesc.HardwareInfo.MemoryDesc,
		CpuCount:    basenode.PeerDesc.HardwareInfo.CpuCount,
		CpuDesc:     basenode.PeerDesc.HardwareInfo.CpuDesc,
		HddCount:    basenode.PeerDesc.HardwareInfo.HddCount,
		HddDesc:     basenode.PeerDesc.HardwareInfo.HddDesc,
	}
}

func (h *PeerHardware) UpdatePeerInfo() error {
	nvmes, _ := runtime.GetNvmeCount()
	nvmeDesc, _ := runtime.GetNvmeDesc()

	gpus, _ := runtime.GetGpuCount()
	gpuDesc, _ := runtime.GetGpuDesc()

	h.NvmeCount = nvmes
	h.NvmeDesc = nvmeDesc

	return nil
}
