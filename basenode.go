package main

import (
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	runtime "github.com/NpoolDevOps/fbc-devops-client/runtime"
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
	NvmeCount int `json:"nvme_should_count"`
	GpuCount  int `json:"gpu_should_count"`
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
	PeerConfig *PeerConfig
}

const peerReportAPI = "https://report.npool.top"

func NewBasenode(config *BasenodeConfig) *Basenode {
	basenode := &Basenode{
		PeerDesc: &PeerDesc{
			PeerConfig: config.PeerConfig,
		},
	}

	basenode.DevopsClient = NewDevopsClient(&DevopsConfig{
		PeerReportAPI: peerReportAPI,
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

	basenode.DevopsClient.FeedMsg(basenode.PeerDesc)

	return basenode
}
