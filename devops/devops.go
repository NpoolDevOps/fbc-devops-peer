package devops

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/node"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"time"
)

type DevopsMsg struct {
	Api   string
	Msg   interface{}
	Retry bool
}

type DevopsConfig struct {
	PeerReportAPI string
	TestMode      bool
}

type DevopsClient struct {
	config *DevopsConfig
	newMsg chan *DevopsMsg
	node   node.Node
}

func NewDevopsClient(config *DevopsConfig) *DevopsClient {
	cli := &DevopsClient{
		config: config,
		newMsg: make(chan *DevopsMsg, 10),
	}

	go cli.reporter()

	return cli
}

func (c *DevopsClient) SetNode(node node.Node) {
	c.node = node
}

func (c *DevopsClient) onMessage(msg *DevopsMsg) {
	if c.config.TestMode {
		log.Infof(log.Fields{}, "runnint in TEST MODE, do not send message")
		return
	}
	b, _ := json.Marshal(msg.Msg)
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(b).
		Post(fmt.Sprintf("%v%v", c.config.PeerReportAPI, msg.Api))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to report message: %v", err)
		if msg.Retry {
			go func() {
				time.Sleep(10 * time.Second)
				c.newMsg <- msg
			}()
		}
		return
	}
	if resp.StatusCode() != 200 {
		if msg.Retry {
			go func() {
				time.Sleep(10 * time.Second)
				c.newMsg <- msg
			}()
		}
		return
	}

	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to report my config: %v", err)
		if msg.Retry {
			go func() {
				time.Sleep(2 * time.Minute)
				c.newMsg <- msg
			}()
		}
		return
	}

	if apiResp.Code != 0 {
		log.Errorf(log.Fields{}, "fail to report my config: %v", apiResp.Msg)
		if msg.Retry {
			go func() {
				time.Sleep(2 * time.Minute)
				c.newMsg <- msg
			}()
		}
		return
	}

	if msg.Api == types.DeviceRegisterAPI {
		b, _ := json.Marshal(apiResp.Body)
		id := types.DeviceRegisterOutput{}
		json.Unmarshal(b, &id)
		c.node.NotifyPeerId(id.Id)
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

func (c *DevopsClient) FeedMsg(api string, msg interface{}, retry bool) {
	go func() {
		c.newMsg <- &DevopsMsg{
			Api:   api,
			Msg:   msg,
			Retry: retry,
		}
	}()
}
