package main

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	runtime "github.com/NpoolDevOps/fbc-devops-client/runtime"
	types "github.com/NpoolDevOps/fbc-devops-service/types"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"time"
)

type Basenode struct {
	DevopsClient   *DevopsClient
	PeerDesc       *PeerDesc
	Owner          string
	Id             uuid.UUID
	PeerConnection *PeerConnection
}

const (
	FullNode      = "fullnode"
	MinerNode     = "miner"
	FullMinerNode = "fullminer"
	WorkerNode    = "worker"
	StorageNode   = "storage"
)

type PeerHardware struct {
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

type PeerDesc struct {
	MySpec       string        `json:"my_spec"`
	HardwareInfo *PeerHardware `json:"hardware_info"`
	PeerConfig   *PeerConfig   `json:"peer_config"`
}

type PeerConfig struct {
	MainRole       string        `json:"main_role"`
	SubRole        string        `json:"sub_role"`
	ParentSpec     string        `json:"parent_spec"`
	HardwareConfig *PeerHardware `json:"hardware_config"`
}

type BasenodeConfig struct {
	PeerReportAPI string
	PeerConfig    *PeerConfig
	Owner         string
}

func NewBasenode(config *BasenodeConfig) *Basenode {
	basenode := &Basenode{
		PeerDesc: &PeerDesc{
			PeerConfig: config.PeerConfig,
		},
		Owner: config.Owner,
	}

	basenode.DevopsClient = NewDevopsClient(&DevopsConfig{
		PeerReportAPI: config.PeerReportAPI,
	})
	if basenode.DevopsClient == nil {
		log.Errorf(log.Fields{}, "fail to create devops client")
		return nil
	}

	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	basenode.PeerDesc.MySpec = spec.SN()

	basenode.PeerDesc.HardwareInfo = &PeerHardware{}
	basenode.PeerDesc.HardwareInfo.UpdatePeerInfo()

	basenode.PeerConnection = NewPeerConnection()
	if basenode.PeerConnection == nil {
		return nil
	}

	basenode.GenerateUuid()

	basenode.PeerConnection.Run()
	basenode.DevopsClient.FeedMsg(types.DeviceRegisterAPI, basenode.ToDeviceRegisterInput())

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
		Spec:        n.PeerDesc.MySpec,
		ParentSpec:  n.PeerDesc.PeerConfig.ParentSpec,
		Role:        n.PeerDesc.PeerConfig.MainRole,
		SubRole:     n.PeerDesc.PeerConfig.SubRole,
		Owner:       n.Owner,
		NvmeCount:   n.PeerDesc.HardwareInfo.NvmeCount,
		NvmeDesc:    n.PeerDesc.HardwareInfo.NvmeDesc,
		GpuCount:    n.PeerDesc.HardwareInfo.GpuCount,
		GpuDesc:     n.PeerDesc.HardwareInfo.GpuDesc,
		MemoryCount: n.PeerDesc.HardwareInfo.MemoryCount,
		MemorySize:  n.PeerDesc.HardwareInfo.MemorySize,
		MemoryDesc:  n.PeerDesc.HardwareInfo.MemoryDesc,
		CpuCount:    n.PeerDesc.HardwareInfo.CpuCount,
		CpuDesc:     n.PeerDesc.HardwareInfo.CpuDesc,
		HddCount:    n.PeerDesc.HardwareInfo.HddCount,
		HddDesc:     n.PeerDesc.HardwareInfo.HddDesc,
	}
}

func (h *PeerHardware) UpdatePeerInfo() error {
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

func (n *Basenode) ReportParentSpec(parentIpParser func() (string, error)) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			var parentIp string
			var err error

			for {
				if parentIpParser != nil {
					parentIp, err = parentIpParser()
					if err != nil {
						<-ticker.C
						continue
					}
				}
				break
			}

			var parentSpec string

			for {
				var spec string
				var err error

				if parentIpParser != nil {
					spec, err = n.PeerConnection.GetParentSpec(parentIp)
				} else {
					spec, err = n.PeerConnection.GetNotifiedParentSpec()
				}
				if err != nil || spec == parentSpec {
					<-ticker.C
					continue
				}
				parentSpec = spec
				break
			}

			n.PeerDesc.PeerConfig.ParentSpec = parentSpec
			n.DevopsClient.FeedMsg(types.DeviceRegisterAPI, n.ToDeviceRegisterInput())

			<-ticker.C
			continue
		}
	}()
}
