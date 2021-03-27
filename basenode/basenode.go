package basenode

import (
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	parser "github.com/NpoolDevOps/fbc-devops-peer/parser"
	"github.com/NpoolDevOps/fbc-devops-peer/peer"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	lic "github.com/NpoolDevOps/fbc-license"
	"github.com/google/uuid"
	"github.com/xjh22222228/ip"
	"golang.org/x/xerrors"
	"net"
	"os/exec"
	"strings"
	"time"
)

type Basenode struct {
	NodeDesc      *NodeDesc
	Username      string
	Password      string
	NetworkType   string
	Id            uuid.UUID
	devopsClient  *devops.DevopsClient
	parser        *parser.Parser
	HasId         bool
	TestMode      bool
	Peer          *peer.Peer
	hasPublicAddr bool
	hasLocalAddr  bool
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
	OsSpec         string        `json:"os_spec"`
}

type BasenodeConfig struct {
	NodeReportAPI string
	NodeConfig    *NodeConfig
	Username      string
	Password      string
	NetworkType   string
	TestMode      bool
}

func NewBasenode(config *BasenodeConfig, devopsClient *devops.DevopsClient) *Basenode {
	basenode := &Basenode{
		NodeDesc: &NodeDesc{
			NodeConfig: config.NodeConfig,
		},
		Username:     config.Username,
		Password:     config.Password,
		NetworkType:  config.NetworkType,
		devopsClient: devopsClient,
		TestMode:     config.TestMode,
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	basenode.NodeDesc.MySpec = spec.SN()

	basenode.NodeDesc.HardwareInfo = &NodeHardware{}
	basenode.NodeDesc.HardwareInfo.UpdateNodeInfo()

	basenode.parser = parser.NewParser()
	basenode.GetAddress()
	basenode.ReadOsSpec()

	role, err := basenode.parser.GetSubRole(basenode.GetMainRole())
	if err == nil {
		basenode.NodeDesc.NodeConfig.SubRole = role
	}

	basenode.startLicenseChecker()
	basenode.devopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput(), true)

	devopsClient.SetNode(basenode)

	return basenode
}

func (n *Basenode) SetPeer(p interface{}) {
	n.Peer = p.(*peer.Peer)
}

func (n *Basenode) Heartbeat(childPeer string) error {
	return n.Peer.Heartbeat(childPeer)
}

func (n *Basenode) MyPublicAddr() (string, error) {
	if !n.hasPublicAddr {
		return "", xerrors.Errorf("public address not validate")
	}
	return n.NodeDesc.NodeConfig.PublicAddr, nil
}

func (n *Basenode) MyLocalAddr() (string, error) {
	if !n.hasLocalAddr {
		return "", xerrors.Errorf("local address not validate")
	}
	return n.NodeDesc.NodeConfig.LocalAddr, nil
}

func (n *Basenode) startLicenseChecker() {
	if !n.TestMode {
		go lic.LicenseChecker(n.Username, n.Password, false, n.NetworkType)
	}

}

func (n *Basenode) ReadOsSpec() {
	out, _ := exec.Command("uname -a").Output()
	n.NodeDesc.NodeConfig.OsSpec = string(out)
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
					n.hasLocalAddr = true
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
					n.hasPublicAddr = true
					updated = true
				}
			}
			if updated {
				n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput(), true)
			}
			<-ticker.C
		}
	}()
}

func (n *Basenode) ToDeviceRegisterInput() *types.DeviceRegisterInput {
	return &types.DeviceRegisterInput{
		Id:          n.Id,
		Spec:        n.NodeDesc.MySpec,
		ParentSpec:  n.NodeDesc.NodeConfig.ParentSpec,
		Role:        n.NodeDesc.NodeConfig.MainRole,
		SubRole:     n.NodeDesc.NodeConfig.SubRole,
		CurrentUser: n.Username,
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
		OsSpec:      n.NodeDesc.NodeConfig.OsSpec,
	}
}

func (h *NodeHardware) UpdateNodeInfo() error {
	nvmes, _ := runtime.GetNvmeCount()
	nvmeDesc, _ := runtime.GetNvmeDesc()

	gpus, _ := runtime.GetGpuCount()
	gpuDesc, _ := runtime.GetGpuDesc()

	mems, _ := runtime.GetMemoryCount()
	memSize, _ := runtime.GetMemorySize()
	memDesc, _ := runtime.GetMemoryDesc()

	cpus, _ := runtime.GetCpuCount()
	cpuDesc, _ := runtime.GetCpuDesc()

	hdds, _ := runtime.GetHddCount()
	hddDesc, _ := runtime.GetHddDesc()

	h.NvmeCount = nvmes
	h.NvmeDesc = nvmeDesc

	h.GpuCount = gpus
	h.GpuDesc = gpuDesc

	h.MemoryCount = mems
	h.MemorySize = memSize
	h.MemoryDesc = memDesc

	h.CpuCount = cpus
	h.CpuDesc = cpuDesc

	h.HddCount = hdds
	h.HddDesc = hddDesc

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
	n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput(), true)
}

func (n *Basenode) GetParentIP() (string, error) {
	return n.parser.GetParentIP(n.GetMainRole())
}

func (n *Basenode) GetChildsIPs() ([]string, error) {
	return n.parser.GetChildsIPs(n.GetMainRole())
}

func (n *Basenode) NotifyPeerId(id uuid.UUID) {
	n.Id = id
	n.HasId = true
}

func (n *Basenode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      BBBBBBBBBassssssseeeNNN      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
