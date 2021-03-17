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
	FullNode    = "fullnode"
	MinerNode   = "miner"
	WorkerNode  = "worker"
	StorageNode = "storage"
)

type PeerHardware struct {
	NvmeCount int `json:"nvme_count"`
	GpuCount  int `json:"gpu_count"`
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

	nvmes, _ := runtime.GetNvmeCount()
	gpus, _ := runtime.GetGpuCount()

	basenode.PeerDesc.HardwareInfo = &PeerHardware{
		NvmeCount: nvmes,
		GpuCount:  gpus,
	}

	basenode.DevopsClient.FeedMsg(types.DeviceRegisterAPI, types.DeviceRegisterInput {
	Spec        string    `gorm:"column:spec" json:"spec"`
	ParentSpec  string    `gorm:"column:parent_spec" json:"parent_spec"`
	Role        string    `gorm:"column:role" json:"role"`
	SubRole     string    `gorm:"column:sub_role" json:"sub_role"`
	Owner       string    `gorm:"column:owner" json:"owner"`
	CurrentUser string    `gorm:"column:current_user" json:"current_user"`
	Manager     string    `gorm:"column:manager" json:"manager"`
	NvmeCount   int       `gorm:"column:nvme_count" json:"nvme_count"`
	NvmeDesc    string    `gorm:"column:nvme_desc" json:"nvme_desc"`
	GpuCount    int       `gorm:"column:gpu_count" json:"gpu_count"`
	GpuDesc     string    `gorm:"column:gpu_desc" json:"gpu_desc"`
	MemoryCount int       `gorm:"column:memory_count" json:"memory_count"`
	MemorySize  uint64    `gorm:"column:memory_size" json:"memory_size"`
	MemoryDesc  string    `gorm:"column:memory_desc" json:"memory_desc"`
	CpuCount    int       `gorm:"column:cpu_count" json:"cpu_count"`
	CpuDesc     string    `gorm:"column:cpu_desc" json:"cpu_desc"`
	HddCount    int       `gorm:"column:hdd_count" json:"hdd_count"`
	HddDesc     string    `gorm:"column:hdd_desc" json:"hdd_desc"`
	})

	return basenode
}
