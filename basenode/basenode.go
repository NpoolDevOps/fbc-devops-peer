package basenode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	basemetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/basemetrics"
	parser "github.com/NpoolDevOps/fbc-devops-peer/parser"
	"github.com/NpoolDevOps/fbc-devops-peer/peer"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	version "github.com/NpoolDevOps/fbc-devops-peer/version"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	lic "github.com/NpoolDevOps/fbc-license"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/xerrors"
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
	addrNotifier  func(string, string)
	BaseMetrics   *basemetrics.BaseMetrics
	Versions      []version.Version
}

type NodeHardware struct {
	NvmeCount     int      `json:"nvme_count"`
	NvmeDesc      []string `json:"nvme_desc"`
	GpuCount      int      `json:"gpu_count"`
	GpuDesc       []string `json:"gpu_desc"`
	MemoryCount   int      `json:"memory_count"`
	MemorySize    uint64   `json:"memory_size"`
	MemoryDesc    []string `json:"memory_desc"`
	CpuCount      int      `json:"cpu_count"`
	CpuDesc       []string `json:"cpu_desc"`
	HddCount      int      `json:"hdd_count"`
	HddDesc       []string `json:"hdd_desc"`
	EthernetCount int      `json:"ethernet_count"`
	EthernetDesc  []string `json:"ethernet_desc"`
}

type NodeDesc struct {
	MySpec       string        `json:"my_spec"`
	HardwareInfo *NodeHardware `json:"hardware_info"`
	NodeConfig   *NodeConfig   `json:"peer_config"`
}

type NodeConfig struct {
	MainRole       string        `json:"main_role"`
	SubRole        string        `json:"sub_role"`
	ParentSpec     []string      `json:"parent_spec"`
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
		hasLocalAddr: true,
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	basenode.NodeDesc.MySpec = spec.SN()

	basenode.NodeDesc.HardwareInfo = &NodeHardware{}
	basenode.NodeDesc.HardwareInfo.UpdateNodeInfo()

	basenode.parser = parser.NewParser()

	basenode.GetAddress()
	basenode.AddressUpdater()

	basenode.ReadOsSpec()

	role, err := basenode.parser.GetSubRole(basenode.GetMainRole())
	if err == nil {
		basenode.NodeDesc.NodeConfig.SubRole = role
	}

	basenode.BaseMetrics = basemetrics.NewBaseMetrics()

	basenode.startLicenseChecker()
	basenode.devopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput(), true)

	devopsClient.SetNode(basenode)

	return basenode
}

func (n *Basenode) SetPeer(p interface{}) {
	n.Peer = p.(*peer.Peer)
}

func (n *Basenode) SetAddrNotifier(addrNotifier func(string, string)) {
	n.addrNotifier = addrNotifier
	var err error
	localAddr, err := n.MyLocalAddr()
	if err != nil {
		return
	}
	publicAddr, err := n.MyPublicAddr()
	if err != nil {
		return
	}
	addrNotifier(localAddr, publicAddr)
}

func (n *Basenode) WatchVersions(localAddr string, err error, versionGetter func(string) []version.Version) {
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		vers := []version.Version{}
		for {
			if err != nil {
				<-ticker.C
				continue
			}
			vs := versionGetter(localAddr)
			updated := false
			for _, ver := range vers {
				for _, v := range vs {
					if v.Application == ver.Application {
						if v.Version != ver.Version {
							updated = true
						}
					}
				}
			}

			if updated {
				n.Versions = vs
				n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput(), true)
			}

			<-ticker.C
		}
	}()
}

func (n *Basenode) SetApplicationVersions(versions []version.Version) {
	n.Versions = versions
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
	out, _ := exec.Command("uname", "-a").Output()
	n.NodeDesc.NodeConfig.OsSpec = string(out)
}

func (n *Basenode) getPublicAddr(url string) (string, error) {
	localAddr := n.NodeDesc.NodeConfig.LocalAddr

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				localAddr, err := net.ResolveTCPAddr(network, fmt.Sprintf("%v:0", localAddr))
				if err != nil {
					return nil, err
				}
				remoteAddr, err := net.ResolveTCPAddr(network, addr)
				if err != nil {
					return nil, err
				}
				conn, err := net.DialTCP(network, localAddr, remoteAddr)
				if err != nil {
					return nil, err
				}
				deadline := time.Now().Add(35 * time.Second)
				conn.SetDeadline(deadline)
				return conn, nil
			},
		},
	}

	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.93 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	addr := net.ParseIP(string(body))
	if addr == nil {
		return "", xerrors.Errorf("invalid ip address")
	}

	return string(body), nil
}

func (n *Basenode) GetAddress() (string, string, error) {
	localAddr := n.NodeDesc.NodeConfig.LocalAddr

	addr, err := exec.Command(
		"dig", "+short", "myip.opendns.com",
		"@resolver1.opendns.com", "-b", localAddr,
	).Output()
	if err == nil {
		n.hasPublicAddr = true
		return localAddr, string(addr), nil
	}

	log.Errorf(log.Fields{}, "cannot get public address with dig: %v", err)

	publicAddr, err := n.getPublicAddr("http://inet-ip.info/ip")
	if err == nil {
		n.hasPublicAddr = true
		return localAddr, publicAddr, err
	}

	log.Errorf(log.Fields{}, "cannot get public address: %v", err)

	publicAddr, err = n.getPublicAddr("http://ipinfo.io/ip")
	if err == nil {
		n.hasPublicAddr = true
		return localAddr, publicAddr, err
	}

	log.Errorf(log.Fields{}, "cannot get public address: %v", err)

	return localAddr, "", err
}

func (n *Basenode) AddressUpdater() {
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			updated := false
			localAddr, publicAddr, err := n.GetAddress()
			if err != nil {
				<-ticker.C
				continue
			}

			if n.NodeDesc.NodeConfig.LocalAddr != localAddr {
				log.Infof(log.Fields{}, "local address updated: %v -> %v",
					n.NodeDesc.NodeConfig.LocalAddr, localAddr)
				n.NodeDesc.NodeConfig.LocalAddr = localAddr
				updated = true
			}

			if n.NodeDesc.NodeConfig.PublicAddr != publicAddr {
				log.Infof(log.Fields{}, "public address updated: %v -> %v",
					n.NodeDesc.NodeConfig.PublicAddr, publicAddr)
				n.NodeDesc.NodeConfig.PublicAddr = publicAddr
				updated = true
			}

			if updated {
				n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput(), true)
				if n.addrNotifier != nil {
					n.addrNotifier(localAddr, publicAddr)
				}
			}
			<-ticker.C
		}
	}()
}

func (n *Basenode) ToDeviceRegisterInput() *types.DeviceRegisterInput {
	versions := []string{}

	for _, ver := range n.Versions {
		b, _ := json.Marshal(ver)
		versions = append(versions, string(b))
	}

	parentSpecs := strings.Join(n.NodeDesc.NodeConfig.ParentSpec, ",")

	return &types.DeviceRegisterInput{
		Id:            n.Id,
		Spec:          n.NodeDesc.MySpec,
		ParentSpec:    parentSpecs,
		Role:          n.NodeDesc.NodeConfig.MainRole,
		SubRole:       n.NodeDesc.NodeConfig.SubRole,
		CurrentUser:   n.Username,
		NvmeCount:     n.NodeDesc.HardwareInfo.NvmeCount,
		NvmeDesc:      n.NodeDesc.HardwareInfo.NvmeDesc,
		GpuCount:      n.NodeDesc.HardwareInfo.GpuCount,
		GpuDesc:       n.NodeDesc.HardwareInfo.GpuDesc,
		MemoryCount:   n.NodeDesc.HardwareInfo.MemoryCount,
		MemorySize:    n.NodeDesc.HardwareInfo.MemorySize,
		MemoryDesc:    n.NodeDesc.HardwareInfo.MemoryDesc,
		CpuCount:      n.NodeDesc.HardwareInfo.CpuCount,
		CpuDesc:       n.NodeDesc.HardwareInfo.CpuDesc,
		HddCount:      n.NodeDesc.HardwareInfo.HddCount,
		HddDesc:       n.NodeDesc.HardwareInfo.HddDesc,
		LocalAddr:     n.NodeDesc.NodeConfig.LocalAddr,
		PublicAddr:    n.NodeDesc.NodeConfig.PublicAddr,
		OsSpec:        n.NodeDesc.NodeConfig.OsSpec,
		EthernetCount: n.NodeDesc.HardwareInfo.EthernetCount,
		EthernetDesc:  n.NodeDesc.HardwareInfo.EthernetDesc,
		Versions:      versions,
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

	eths, _ := runtime.GetEthernetCount()
	ethDesc, _ := runtime.GetEthernetDesc()

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

	h.EthernetCount = eths
	h.EthernetDesc = ethDesc

	return nil
}

func (n *Basenode) GetMainRole() string {
	return n.NodeDesc.NodeConfig.MainRole
}

func (n *Basenode) GetSubRole() string {
	return n.NodeDesc.NodeConfig.SubRole
}

func (n *Basenode) NotifyParentSpec(spec string) {
	for _, pspec := range n.NodeDesc.NodeConfig.ParentSpec {
		if pspec == spec {
			return
		}
	}
	n.NodeDesc.NodeConfig.ParentSpec = append(n.NodeDesc.NodeConfig.ParentSpec, spec)
	n.devopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput(), true)
}

func (n *Basenode) GetParentIP() (string, error) {
	return n.parser.GetParentIP(n.GetMainRole())
}

func (n *Basenode) GetChildsIPs() ([]string, error) {
	return n.parser.GetChildsIPs(n.GetMainRole())
}

func (n *Basenode) GetLogFileByRole(role string) (string, error) {
	return n.parser.GetLogFile(role)
}

func (n *Basenode) GetFullnodeHost() (string, error) {
	return n.parser.GetFullnodeHost()
}

func (n *Basenode) GetShareStorageRoot() (string, error) {
	return n.parser.GetShareStorageRoot(n.GetMainRole())
}
func (n *Basenode) GetShareStorageRootByRole(role string) (string, error) {
	return n.parser.GetShareStorageRoot(role)
}

func (n *Basenode) GetLogFile() (string, error) {
	return n.parser.GetLogFile(n.GetMainRole())
}

func (n *Basenode) NotifyPeerId(id uuid.UUID) {
	n.Id = id
	n.HasId = true
}

func (n *Basenode) Describe(ch chan<- *prometheus.Desc) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR BASENODE")
}

func (n *Basenode) Collect(ch chan<- prometheus.Metric) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR BASENODE")
}

func (n *Basenode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}

func (n *Basenode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      BBBBBBBBBassssssseeeNNN      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
