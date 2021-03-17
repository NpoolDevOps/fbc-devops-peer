package main

import (
	"time"
)

type DevopsConfig struct {
	PeerReportAPI string
	PeerConfig    *PeerConfig
}

type DevopsClient struct {
	config   *DevopsConfig
	peerDesc PeerDesc
}

func NewDevopsClient(config *DevopsConfig) *DevopsClient {
	cli := &DevopsClient{
		config: config,
	}

	go cli.reportMySelf()

	return cli
}

func (c *DevopsClient) reportMySelf() {
	ticker := time.NewTicker(3 * time.Minute)
	for {
		<-ticker.C
	}
}
