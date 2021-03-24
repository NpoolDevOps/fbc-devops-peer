package basenode

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	parser "github.com/NpoolDevOps/fbc-devops-peer/parser"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	"github.com/google/uuid"
	"github.com/xjh22222228/ip"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

type Basenode struct {
	NodeDesc     *NodeDesc
	User         string
	Id           uuid.UUID
	devopsClient *devops.DevopsClient
	parser       *parser.Parser
}

type NodeHardware struct {
	NvmeCount   int      `json:"nvme_count"`
	NvmeDesc    []string `json:"nvme_desc"`
	GpuCount    int      `json:"gpu_count"`
	GpuDesc     []string `json:"gpu_desc"`
	MemoryCount int      `json:"memory_count"`
	MemorySize  uint64   `json:"memory_size"`
	MemoryDesc  []string `json:"memory_desc"`
	CpuCount    int      `json:"cpu_count"`
	CpuDesc     []string `json:"cpu_desc"`
	HddCount    int      `json:"hdd_count"`
	HddDesc     []string `json:"hdd_count"`
}

type NodeDesc struct {
	MySpec       string        `json:"my_spec"`
	HardwareInfo *NodeHardware `json:"hardware_info"`
	NodeConfig   *NodeConfig   `json:"peer_config"`
}

type NodeConfig struct {
	MainRole       string        `json:"main_role"`
	SubRole        string        `json:"sub_role"`
	ParentSpec     string        `json:"parent_spec"`
	HardwareConfig *NodeHardware `json:"hardware_config"`
	LocalAddr      string        `json:"local_addr"`
	PublicAddr     string        `json:"public_addr"`
}

type BasenodeConfig struct {
	NodeReportAPI string
	NodeConfig    *NodeConfig
	User          string
}

func NewBasenode(config *BasenodeConfig, devopsClient *devops.DevopsClient) *Basenode {
	basenode := &Basenode{
		NodeDesc: &NodeDesc{
			NodeConfig: config.NodeConfig,
		},
		User:         config.User,
		devopsClient: devopsClient,
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	basenode.NodeDesc.MySpec = spec.SN()

	basenode.NodeDesc.HardwareInfo = &NodeHardware{}
	basenode.NodeDesc.HardwareInfo.UpdateNodeInfo()

	basenode.GenerateUuid()
	basenode.parser = parser.NewParser()
	basenode.GetAddress()

	basenode.devopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput())

	return basenode
}

func (n *Basenode) GetAddress() {
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			updated := false

			conn, err := net.Dial("udp", "8.8.8.8:80")
			if err == nil {
				localAddr := strings.Split(conn.LocalAddr().String(), ":")[0]
				if n.NodeDesc.NodeConfig.LocalAddr != localAddr {
					log.Infof(log.Fields{}, "local address updated: %v -> %v",
						n.NodeDesc.NodeConfig.LocalAddr, localAddr)
					n.NodeDesc.NodeConfig.LocalAddr = localAddr
					updated = true
				}
				conn.Close()
			}
			publicAddr, err := ip.V4()
			if err == nil {
				if n.NodeDesc.NodeConfig.PublicAddr != publicAddr {
					log.Infof(log.Fields{}, "public address updated: %v -> %v",
						n.NodeDesc.NodeConfig.PublicAddr, publicAddr)
					n.NodeDesc.NodeConfig.PublicAddr = publicAddr
					n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput())
					updated = true
				}
			}
			if updated {
				n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput())
			}
			<-ticker.C
		}
	}()
}

type BasenodeConfigFromFile struct {
	Id uuid.UUID `json:"id"`
}

func (n *Basenode) GenerateUuid() {
	var cfgFromFile BasenodeConfigFromFile

	env := os.Getenv("HOME")

	uuidPath := fmt.Sprintf("%s/.fbc-devops-peer", env)
	uuidFile := fmt.Sprintf("%s/peer.conf", uuidPath)
	b, err := ioutil.ReadFile(uuidFile)
	if err == nil {
		err = json.Unmarshal(b, &cfgFromFile)
		if err == nil {
			n.Id = cfgFromFile.Id
			return
		}
	}
	cfgFromFile.Id = uuid.New()
	b, err = json.Marshal(cfgFromFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse config to json")
	}

	os.MkdirAll(uuidPath, 0755)
	err = ioutil.WriteFile(uuidFile, b, 0644)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot write uuid file: %v", err)
	}
	n.Id = cfgFromFile.Id
}

func (n *Basenode) ToDeviceRegisterInput() *types.DeviceRegisterInput {
	return &types.DeviceRegisterInput{
		Id:          n.Id,
		Spec:        n.NodeDesc.MySpec,
		ParentSpec:  n.NodeDesc.NodeConfig.ParentSpec,
		Role:        n.NodeDesc.NodeConfig.MainRole,
		SubRole:     n.NodeDesc.NodeConfig.SubRole,
		CurrentUser: n.User,
		NvmeCount:   n.NodeDesc.HardwareInfo.NvmeCount,
		NvmeDesc:    n.NodeDesc.HardwareInfo.NvmeDesc,
		GpuCount:    n.NodeDesc.HardwareInfo.GpuCount,
		GpuDesc:     n.NodeDesc.HardwareInfo.GpuDesc,
		MemoryCount: n.NodeDesc.HardwareInfo.MemoryCount,
		MemorySize:  n.NodeDesc.HardwareInfo.MemorySize,
		MemoryDesc:  n.NodeDesc.HardwareInfo.MemoryDesc,
		CpuCount:    n.NodeDesc.HardwareInfo.CpuCount,
		CpuDesc:     n.NodeDesc.HardwareInfo.CpuDesc,
		HddCount:    n.NodeDesc.HardwareInfo.HddCount,
		HddDesc:     n.NodeDesc.HardwareInfo.HddDesc,
		LocalAddr:   n.NodeDesc.NodeConfig.LocalAddr,
		PublicAddr:  n.NodeDesc.NodeConfig.PublicAddr,
	}
}

func (h *NodeHardware) UpdateNodeInfo() error {
	nvmes, _ := runtime.GetNvmeCount()
	nvmeDesc, _ := runtime.GetNvmeDesc()

	gpus, _ := runtime.GetGpuCount()
	gpuDesc, _ := runtime.GetGpuDesc()

	h.NvmeCount = nvmes
	h.NvmeDesc = nvmeDesc

	h.GpuCount = gpus
	h.GpuDesc = gpuDesc

	return nil
}

func (n *Basenode) GetMainRole() string {
	return n.NodeDesc.NodeConfig.MainRole
}

func (n *Basenode) GetSubRole() string {
	return n.NodeDesc.NodeConfig.SubRole
}

func (n *Basenode) NotifyParentSpec(spec string) {
	if n.NodeDesc.NodeConfig.ParentSpec == spec {
		return
	}
	n.NodeDesc.NodeConfig.ParentSpec = spec
	n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput())
}

func (n *Basenode) GetParentIP() (string, error) {
	return n.parser.GetParentIP(n.GetMainRole())
}
