package main

import (
	"time"
)

type DevopsConfig struct {
	PeerReportAPI string
}

type DevopsClient struct {
	config *DevopsConfig
	newMsg chan string
}

func NewDevopsClient(config *DevopsConfig) *DevopsClient {
	cli := &DevopsClient{
		config: config,
		newMsg: make(chan string, 10),
	}

	go cli.reporter()

	return cli
}

func (c *DevopsClient) reporter() {
	ticker := time.NewTicker(3 * time.Minute)
	for {
		<-ticker.C
	}
}

func (c *DevopsClient) FeedMsg(msg string) {
	go func() { c.newMsg <- msg }()
}
