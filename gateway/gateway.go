package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	mytypes "github.com/NpoolDevOps/fbc-devops-peer/types"
	devopsapi "github.com/NpoolDevOps/fbc-devops-service/devopsapi"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	"github.com/NpoolRD/http-daemon"
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
	localAddr  string
	newCreated bool
}

type GatewayNode struct {
	*basenode.Basenode
	topologyTicker  *time.Ticker
	addressWaiter   chan struct{}
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
		case <-g.addressWaiter:
			g.waitForAddr()
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
			localAddr:  device.LocalAddr,
			online:     online,
			newCreated: newCreated,
		}
		if device.Role == mytypes.StorageNode {
			if device.SubRole == mytypes.StorageRoleMgr {
				monitor.ports = append(monitor.ports, 9283)
			}
		}
		if device.Role != mytypes.StorageNode && device.Role != mytypes.FullNode {
			monitor.ports = append(monitor.ports, 9400)
		}
		g.hosts[device.LocalAddr] = monitor
	}

	go func() { g.addressWaiter <- struct{}{} }()
}

func (g *GatewayNode) waitForAddr() {
	_, err := g.MyPublicAddr()
	if err != nil {
		log.Errorf(log.Fields{}, "public address is not ready: %v", err)
		time.Sleep(10 * time.Second)
		go func() { g.addressWaiter <- struct{}{} }()
		return
	}

	_, err = g.MyLocalAddr()
	if err != nil {
		log.Errorf(log.Fields{}, "local address is not ready: %v", err)
		time.Sleep(10 * time.Second)
		go func() { g.addressWaiter <- struct{}{} }()
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

type staticConfig struct {
	Targets []string `yaml:"targets"`
}

type scrapeConfig struct {
	JobName        string         `yaml:"job_name"`
	ScrapeInterval string         `yaml:"scrape_interval,omitempty"`
	StaticConfigs  []staticConfig `yaml:"static_configs"`
}

type alertManager struct {
	StaticConfigs []staticConfig `yaml:"static_configs"`
}

type alerting struct {
	AlertManagers []alertManager `yaml:"alertmanagers"`
}

type monitorGlobal struct {
	ScrapeInterval     string `yaml:"scrape_interval"`
	ScrapeTimeout      string `yaml:"scrape_timeout"`
	EvaluationInterval string `yaml:"evaluation_interval"`
}

type monitorConfig struct {
	Global        monitorGlobal  `yaml:"global"`
	Alerting      alerting       `yaml:"alerting"`
	RuleFiles     []string       `yaml:"rule_files"`
	ScrapeConfigs []scrapeConfig `yaml:"scrape_configs"`
}

func (g *GatewayNode) generateConfig() {
	monitorCfgPath := filepath.Join(os.Getenv("HOME"), ".fbc-devops-peer")
	os.MkdirAll(monitorCfgPath, 0755)
	monitorCfgFile := filepath.Join(monitorCfgPath, "fbc-peer-monitor.yml")

	exec.Command("rm", "-rf", monitorCfgFile).Run()
	myLocalAddr, _ := g.MyLocalAddr()

	config := monitorConfig{
		Global: monitorGlobal{
			ScrapeInterval:     "1m",
			ScrapeTimeout:      "50s",
			EvaluationInterval: "1m",
		},
		Alerting: alerting{
			AlertManagers: []alertManager{
				{
					StaticConfigs: []staticConfig{
						{
							Targets: []string{
								"localhost:9093",
								"alertmanager.npool.top",
								// TODO: add customized alert url
							},
						},
					},
				},
			},
		},
		RuleFiles: []string{
			"rules/*",
			// TODO: try to get rule files from cloud
		},
		ScrapeConfigs: []scrapeConfig{
			{
				JobName: "prometheus",
				StaticConfigs: []staticConfig{
					{
						Targets: []string{
							fmt.Sprintf("%s:%v", myLocalAddr, 9090),
						},
					},
				},
			},
		},
	}

	roleHostMap := map[string][]hostMonitor{}

	for _, monitor := range g.hosts {
		roleHostMap[monitor.role] = append(roleHostMap[monitor.role], monitor)
	}

	for role, monitors := range roleHostMap {
		jobConfig := scrapeConfig{
			JobName: role,
		}

		subConfigs := []staticConfig{}
		targets := []string{}

		for _, monitor := range monitors {
			for _, port := range monitor.ports {
				targets = append(targets, fmt.Sprintf("%v:%v", monitor.localAddr, port))
			}
		}
		subConfigs = append(subConfigs, staticConfig{
			Targets: targets,
		})
		jobConfig.StaticConfigs = subConfigs
		config.ScrapeConfigs = append(config.ScrapeConfigs, jobConfig)
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

	exec.Command("mv", monitorCfgFile, "/usr/local/prometheus/prometheus.yml")
	g.reloadConfig()
}

func (g *GatewayNode) reloadConfig() {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		Post(fmt.Sprintf("http://localhost:9090/-/reload"))
	if err != nil {
		log.Errorf(log.Fields{}, "cannot reload monitor config")
		return
	}
	if resp.StatusCode() != 200 {
		log.Errorf(log.Fields{}, "fail to reload monitor config")
	}
}

func (g *GatewayNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      GGGGGGattttttEEEEWayyyy      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
