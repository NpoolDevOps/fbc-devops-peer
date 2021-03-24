package basenode

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	runtime "github.com/NpoolDevOps/fbc-devops-peer/runtime"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
)

type Basenode struct {
	NodeDesc *NodeDesc
	User     string
	Id       uuid.UUID
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
}

type BasenodeConfig struct {
	NodeReportAPI string
	NodeConfig    *NodeConfig
	User          string
}

func NewBasenode(config *BasenodeConfig) *Basenode {
	basenode := &Basenode{
		NodeDesc: &NodeDesc{
			NodeConfig: config.NodeConfig,
		},
		User: config.User,
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	basenode.NodeDesc.MySpec = spec.SN()

	basenode.NodeDesc.HardwareInfo = &NodeHardware{}
	basenode.NodeDesc.HardwareInfo.UpdateNodeInfo()

	basenode.GenerateUuid()

	return basenode
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
	log.Infof(log.Fields{}, "Parent SPEC notified: %v", spec)
}
