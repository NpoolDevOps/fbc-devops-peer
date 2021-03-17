package main

import (
	machspec "github.com/EntropyPool/machine-spec"
	_ "github.com/NpoolDevOps/fbc-devops-client/runtime"
	"time"
)

type DevopsConfig struct {
	PeerReportAPI string
	PeerConfig    *PeerConfig
}

type DevopsClient struct {
	config     *DevopsConfig
	peerDesc   PeerDesc
	peerErrors []error
}

func NewDevopsClient(config *DevopsConfig) *DevopsClient {
	cli := &DevopsClient{
		config: config,
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	cli.peerDesc.MySpec = spec.SN()

	go cli.reportMySelf()

	return cli
}

func (c *DevopsClient) reportMySelf() {
	ticker := time.NewTicker(3 * time.Minute)
	for {
		<-ticker.C
	}
}
