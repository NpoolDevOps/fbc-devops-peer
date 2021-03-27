package gateway

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
)

type GatewayNode struct {
	*basenode.Basenode
}

func NewGatewayNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *GatewayNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	gateway := &GatewayNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return gateway
}

func (g *GatewayNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      GGGGGGattttttEEEEWayyyy      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
