package gateway

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	devopsapi "github.com/NpoolDevOps/fbc-devops-service/devopsapi"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	"time"
)

type GatewayNode struct {
	*basenode.Basenode
	topologyTicker  *time.Ticker
	configGenerator chan struct{}
	lastTopologySum string
}

func NewGatewayNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *GatewayNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	gateway := &GatewayNode{
		basenode.NewBasenode(config, devopsClient),
		time.NewTicker(2 * time.Minute),
		make(chan struct{}, 10),
		"",
	}

	gateway.updateTopology()
	go gateway.handler()

	return gateway
}

func (g *GatewayNode) handler() {
	for {
		select {
		case <-g.topologyTicker.C:
			g.updateTopology()
		case <-g.configGenerator:
			g.generateConfig()
		}
	}
}

func (g *GatewayNode) updateTopology() {
	output, err := devopsapi.MyDevicesByUsername(types.MyDevicesByUsernameInput{
		Username: g.Username,
		Password: g.Password,
	})
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get devices by username: %v", err)
		return
	}

	log.Infof(log.Fields{}, "output -> %v", output)
}

func (g *GatewayNode) generateConfig() {

}

func (g *GatewayNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      GGGGGGattttttEEEEWayyyy      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
