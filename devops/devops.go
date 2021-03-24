package devops

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"time"
)

type DevopsMsg struct {
	Api string
	Msg interface{}
}

type DevopsConfig struct {
	PeerReportAPI string
}

type DevopsClient struct {
	config *DevopsConfig
	newMsg chan *DevopsMsg
}

func NewDevopsClient(config *DevopsConfig) *DevopsClient {
	cli := &DevopsClient{
		config: config,
		newMsg: make(chan *DevopsMsg, 10),
	}

	go cli.reporter()

	return cli
}

func (c *DevopsClient) onMessage(msg *DevopsMsg) {
	b, _ := json.Marshal(msg.Msg)
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(b).
		Post(fmt.Sprintf("%v%v", c.config.PeerReportAPI, msg.Api))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to report message: %v", err)
		go func() {
			time.Sleep(10 * time.Second)
			c.newMsg <- msg
		}()
		return
	}
	if resp.StatusCode() != 200 {
		go func() {
			time.Sleep(10 * time.Second)
			c.newMsg <- msg
		}()
		return
	}

	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to report my config: %v", err)
		return
	}

	if apiResp.Code != 0 {
		log.Errorf(log.Fields{}, "fail to report my config: %v", apiResp.Msg)
	}
}

func (c *DevopsClient) reporter() {
	ticker := time.NewTicker(3 * time.Minute)
	for {
		select {
		case msg := <-c.newMsg:
			c.onMessage(msg)
		case <-ticker.C:
		}
	}
}

func (c *DevopsClient) FeedMsg(api string, msg interface{}) {
	go func() {
		c.newMsg <- &DevopsMsg{
			Api: api,
			Msg: msg,
		}
	}()
}
