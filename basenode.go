package main

import (
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	runtime "github.com/NpoolDevOps/fbc-devops-client/runtime"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
)

type Basenode struct {
	DevopsClient   *DevopsClient
	PeerDesc       *PeerDesc
	Owner          string
	PeerConnection *PeerConnection
}

const (
	FullNode      = "fullnode"
	MinerNode     = "miner"
	FullMinerNode = "fullminer"
	WorkerNode    = "worker"
	StorageNode   = "storage"
)

type PeerHardware struct {
	NvmeCount   int      `json:"nvme_count"`
	NvmeDesc    []string `json:"nvme_desc"`
	GpuCount    int      `json:"gpu_count"`
	GpuDesc     []string `json:"gpu_desc"`
	MemoryCount int      `json:"memory_count"`
	MemorySize  uint64   `json:"memory_size"`
	MemoryDesc  []string `json:"memory_desc"`
	CpuCount    int      `json:"cpu_count"`
	CpuDesc     []string `json:"cpu_desc"`
	HddCount    int      `json:"hdd_count"`
	HddDesc     []string `json:"hdd_count"`
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
		Owner: config.Owner,
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

	basenode.PeerConnection = NewPeerConnection()
	if basenode.PeerConnection == nil {
		return nil
	}

	basenode.PeerConnection.Run()
	GetParentSpec(basenode.PeerConnection, func(parentSpec string) {
		basenode.PeerDesc.PeerConfig.ParentSpec = parentSpec
		basenode.DevopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput())
	})

	basenode.DevopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput())

	return basenode
}

func (n *Basenode) ToDeviceRegisterInput() *types.DeviceRegisterInput {
	return &types.DeviceRegisterInput{
		Spec:        n.PeerDesc.MySpec,
		ParentSpec:  n.PeerDesc.PeerConfig.ParentSpec,
		Role:        n.PeerDesc.PeerConfig.MainRole,
		SubRole:     n.PeerDesc.PeerConfig.SubRole,
		Owner:       n.Owner,
		NvmeCount:   n.PeerDesc.HardwareInfo.NvmeCount,
		NvmeDesc:    n.PeerDesc.HardwareInfo.NvmeDesc,
		GpuCount:    n.PeerDesc.HardwareInfo.GpuCount,
		GpuDesc:     n.PeerDesc.HardwareInfo.GpuDesc,
		MemoryCount: n.PeerDesc.HardwareInfo.MemoryCount,
		MemorySize:  n.PeerDesc.HardwareInfo.MemorySize,
		MemoryDesc:  n.PeerDesc.HardwareInfo.MemoryDesc,
		CpuCount:    n.PeerDesc.HardwareInfo.CpuCount,
		CpuDesc:     n.PeerDesc.HardwareInfo.CpuDesc,
		HddCount:    n.PeerDesc.HardwareInfo.HddCount,
		HddDesc:     n.PeerDesc.HardwareInfo.HddDesc,
	}
}

func (h *PeerHardware) UpdatePeerInfo() error {
	nvmes, _ := runtime.GetNvmeCount()
	nvmeDesc, _ := runtime.GetNvmeDesc()

	gpus, _ := runtime.GetGpuCount()
	gpuDesc, _ := runtime.GetGpuDesc()

	h.NvmeCount = nvmes
	h.NvmeDesc = nvmeDesc

	h.GpuCount = gpus
	h.GpuDesc = gpuDesc

	return nil
}
