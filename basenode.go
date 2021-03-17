package main

import (
	log "github.com/EntropyPool/entropy-logger"
)

type Basenode struct {
	devopsClient *DevopsClient
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
	MySpec       string `json:"my_spec"`
	HardwareInfo PeerHardware
	PeerConfig
}

type PeerConfig struct {
	MainRole   string
	SubRole    string
	ParentSpec string `json:"parent_spec"`
	PeerHardware
}

type BasenodeConfig struct {
	DevopsConfig *DevopsConfig
}

const peerReportAPI = "https://report.npool.top"

func NewBasenode(config *BasenodeConfig) *Basenode {
	basenode := &Basenode{}

	config.DevopsConfig.PeerReportAPI = peerReportAPI

	basenode.devopsClient = NewDevopsClient(config.DevopsConfig)
	if basenode.devopsClient == nil {
		log.Errorf(log.Fields{}, "fail to create devops client")
		return nil
	}

	return basenode
}
