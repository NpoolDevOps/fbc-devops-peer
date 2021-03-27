package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	mytypes "github.com/NpoolDevOps/fbc-devops-peer/types"
	devopsapi "github.com/NpoolDevOps/fbc-devops-service/devopsapi"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	"strings"
	"time"
)

type hostMonitor struct {
	role       string
	ports      []int
	online     bool
	publicAddr string
}

type GatewayNode struct {
	*basenode.Basenode
	topologyTicker  *time.Ticker
	onlineChecker   chan struct{}
	configGenerator chan struct{}
	hosts           map[string]hostMonitor
}

func NewGatewayNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *GatewayNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	gateway := &GatewayNode{
		basenode.NewBasenode(config, devopsClient),
		time.NewTicker(2 * time.Minute),
		make(chan struct{}, 10),
		make(chan struct{}, 10),
		make(map[string]hostMonitor, 0),
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
		case <-g.onlineChecker:
			g.onlineCheck()
		case <-g.configGenerator:
			g.generateConfig()
		}
	}
}

func (g *GatewayNode) updateTopology() {
	passHash := sha256.Sum256([]byte(g.Password))
	output, err := devopsapi.MyDevicesByUsername(types.MyDevicesByUsernameInput{
		Username: g.Username,
		Password: hex.EncodeToString(passHash[0:])[0:12],
	}, true)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get devices by username: %v", err)
		return
	}

	for _, device := range output.Devices {
		online := false

		if _, ok := g.hosts[device.LocalAddr]; ok {
			online = g.hosts[device.LocalAddr].online
		}

		monitor := hostMonitor{
			role:       device.Role,
			ports:      []int{9100, 9256},
			publicAddr: device.PublicAddr,
			online:     online,
		}
		if device.Role == mytypes.StorageNode {
			if device.SubRole == mytypes.StorageRoleMgr {
				monitor.ports = append(monitor.ports, 9283)
			}
		}
		if 0 < device.GpuCount {
			monitor.ports = append(monitor.ports, 9400)
		}
		g.hosts[device.LocalAddr] = monitor
	}

	go func() { g.onlineChecker <- struct{}{} }()
}

func (g *GatewayNode) onlineCheck() {
	myPublicAddr, err := g.MyPublicAddr()
	if err != nil {
		return
	}

	updated := false
	for host, monitor := range g.hosts {
		if host[:strings.LastIndex(host, ".")] != myPublicAddr[:strings.LastIndex(myPublicAddr, ".")] {
			continue
		}

		err := g.Heartbeat(host)
		online := monitor.online
		if err != nil {
			monitor.online = false
		}
		if monitor.online != online {
			updated = true
		}
		g.hosts[host] = monitor
	}

	if updated {
		go func() { g.configGenerator <- struct{}{} }()
	}
}

func (g *GatewayNode) generateConfig() {
	for host, monitor := range g.hosts {
		log.Infof(log.Fields{}, "HOST: %v [%v]", host, monitor.role)
		log.Infof(log.Fields{}, "  PORTS: %v", monitor.ports)
	}
}

func (g *GatewayNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      GGGGGGattttttEEEEWayyyy      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
