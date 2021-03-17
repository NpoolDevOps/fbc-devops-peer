package main

import (
	"encoding/json"
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
		select {
		case <-c.newMsg:
		case <-ticker.C:
		}
	}
}

func (c *DevopsClient) FeedMsg(msg interface{}) {
	b, _ := json.Marshal(msg)
	go func() { c.newMsg <- string(b) }()
}
