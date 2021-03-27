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
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type hostMonitor struct {
	role       string
	ports      []int
	online     bool
	publicAddr string
	newCreated bool
}

type GatewayNode struct {
	*basenode.Basenode
	topologyTicker   *time.Ticker
	publicAddrWaiter chan struct{}
	onlineChecker    chan struct{}
	configGenerator  chan struct{}
	hosts            map[string]hostMonitor
}

func NewGatewayNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *GatewayNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	gateway := &GatewayNode{
		basenode.NewBasenode(config, devopsClient),
		time.NewTicker(2 * time.Minute),
		make(chan struct{}, 10),
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
		case <-g.publicAddrWaiter:
			g.waitForPublicAddr()
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
		newCreated := false

		if _, ok := g.hosts[device.LocalAddr]; !ok {
			newCreated = true
		} else {
			online = g.hosts[device.LocalAddr].online
		}

		monitor := hostMonitor{
			role:       device.Role,
			ports:      []int{9100, 9256},
			publicAddr: device.PublicAddr,
			online:     online,
			newCreated: newCreated,
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

	go func() { g.publicAddrWaiter <- struct{}{} }()
}

func (g *GatewayNode) waitForPublicAddr() {
	_, err := g.MyPublicAddr()
	if err != nil {
		log.Errorf(log.Fields{}, "public address is not ready: %v", err)
		time.Sleep(10 * time.Second)
		go func() { g.publicAddrWaiter <- struct{}{} }()
		return
	}

	go func() { g.onlineChecker <- struct{}{} }()
}

func (g *GatewayNode) onlineCheck() {
	myPublicAddr, _ := g.MyPublicAddr()

	updated := false
	for host, monitor := range g.hosts {
		lastIndex := strings.LastIndex(monitor.publicAddr, ".")
		if lastIndex < 0 {
			log.Errorf(log.Fields{}, "%v miss public address: %v", host, monitor.publicAddr)
			continue
		}

		hostPrefix := monitor.publicAddr[:lastIndex]
		myAddrPrefix := myPublicAddr[:strings.LastIndex(myPublicAddr, ".")]
		if hostPrefix != myAddrPrefix {
			log.Infof(log.Fields{}, "public address prefix %v != %v", hostPrefix, myAddrPrefix)
			continue
		}

		err := g.Heartbeat(host)
		online := monitor.online
		if err != nil {
			log.Infof(log.Fields{}, "heartbeat to %v lost: %v", host, err)
			monitor.online = false
		} else {
			monitor.online = true
		}
		if monitor.online != online {
			updated = true
		}

		if monitor.newCreated {
			updated = true
		}

		g.hosts[host] = monitor
	}

	if updated {
		go func() { g.configGenerator <- struct{}{} }()
	}
}

type monitorGlobal struct {
	ScrapeInterval     string `yaml:"scrape_interval"`
	ScrapeTimeout      string `yaml:"scrape_timeout"`
	EvaluationInterval string `yaml:"evaluation_interval"`
}

type monitorConfig struct {
	Global monitorGlobal `yaml:"global"`
}

func (g *GatewayNode) generateConfig() {
	monitorCfgPath := filepath.Join(os.Getenv("HOME"), ".fbc-devops-peer")
	os.MkdirAll(monitorCfgPath, 0755)
	monitorCfgFile := filepath.Join(monitorCfgPath, "fbc-peer-monitor.yml")

	exec.Command("rm", "-rf", monitorCfgFile).Run()

	config := monitorConfig{
		Global: monitorGlobal{
			ScrapeInterval:     "1m",
			ScrapeTimeout:      "50s",
			EvaluationInterval: "1m",
		},
	}

	b, err := yaml.Marshal(&config)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to marshal config")
		return
	}

	err = ioutil.WriteFile(monitorCfgFile, b, 0755)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to write %v: %v", monitorCfgFile, err)
		return
	}

	// exec.Command("mv", monitorCfgFile, "/usr/local/prometheus/prometheus.yml")
}

func (g *GatewayNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      GGGGGGattttttEEEEWayyyy      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
