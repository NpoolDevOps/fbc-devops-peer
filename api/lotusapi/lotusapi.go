package lotusapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusbase"
	"github.com/NpoolDevOps/fbc-devops-peer/api/minerapi"
	"github.com/NpoolDevOps/fbc-devops-peer/version"
	"github.com/filecoin-project/go-state-types/dline"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/xerrors"
)

type SyncState struct {
	HeightDiff   int64
	BlockElapsed time.Duration
	SyncError    bool
	NetPeers     int
}

func FileWorkerOpened() int64 {
	lotus_pid, err := minerapi.GetDevicePid("lotus")
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get lotus pid", err)
		return 0
	}
	lotus_file_num, err := minerapi.GetDeviceFileOpened(lotus_pid)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get lotus file opened number", err)
		return 0
	}
	num := lotus_file_num
	return num
}

func lotusRpcUrl(host string) string {
	return fmt.Sprintf("http://%v:1234/rpc/v0", host)
}

func stateHeightDiff(state api.SyncState, host string) (int64, error) {
	bh, err := lotusbase.Request(lotusRpcUrl(host), []string{}, "Filecoin.ChainHead")
	if err != nil {
		return -1, err
	}

	head := types.TipSet{}
	json.Unmarshal(bh, &head)

	working := -1
	for i, ss := range state.ActiveSyncs {
		switch ss.Stage {
		case api.StageSyncComplete:
		default:
			working = i
		case api.StageIdle:
			// not complete, not actively working
		}
	}

	if working == -1 {
		working = len(state.ActiveSyncs) - 1
	}

	ss := state.ActiveSyncs[working]
	var heightDiff int64

	if ss.Base != nil {
		heightDiff = int64(ss.Base.Height())
	}
	if ss.Target != nil {
		heightDiff = int64(ss.Target.Height()) - heightDiff
	} else {
		heightDiff = 0
	}

	return heightDiff, nil
}

func stateSyncElapsed(state api.SyncState) (time.Duration, bool) {
	elapsed := 0 * time.Second
	errorHappen := false

	for _, ss := range state.ActiveSyncs {
		e := 0 * time.Second
		if ss.End.IsZero() {
			if !ss.Start.IsZero() {
				e = time.Since(ss.Start)
			}
		} else {
			e = ss.End.Sub(ss.Start)
		}
		if ss.Stage == api.StageSyncErrored {
			errorHappen = true
		}
		if elapsed < e {
			elapsed = e
		}
	}

	return elapsed, errorHappen
}

func ChainSyncState(host string) (*SyncState, error) {
	bs, err := lotusbase.Request(lotusRpcUrl(host), []string{}, "Filecoin.SyncState")
	if err != nil {
		return nil, err
	}

	state := api.SyncState{}
	json.Unmarshal(bs, &state)

	if len(state.ActiveSyncs) == 0 {
		return nil, xerrors.Errorf("no active sync running")
	}

	heightDiff, err := stateHeightDiff(state, host)
	if err != nil {
		return nil, err
	}

	elapsed, errorHappen := stateSyncElapsed(state)

	return &SyncState{
		HeightDiff:   heightDiff,
		BlockElapsed: elapsed,
		SyncError:    errorHappen,
	}, nil
}

func ClientNetPeers(host string) (int, error) {
	bs, err := lotusbase.Request(lotusRpcUrl(host), []string{}, "Filecoin.NetPeers")
	if err != nil {
		return -1, err
	}

	peers := []peer.AddrInfo{}
	json.Unmarshal(bs, &peers)

	return len(peers), nil
}

func ClientVersion(host string) (version.Version, error) {
	bs, err := lotusbase.Request(lotusRpcUrl(host), []string{}, "Filecoin.Version")
	if err != nil {
		return version.Version{}, err
	}

	ver := api.APIVersion{}
	json.Unmarshal(bs, &ver)

	return version.Version{
		Application: "lotus",
		Version:     ver.Version,
	}, nil
}

func TipSetByHeight(host string, height uint64) ([]string, error) {
	bs, err := lotusbase.Request(lotusRpcUrl(host), []interface{}{height, nil}, "Filecoin.ChainGetTipSetByHeight")
	if err != nil {
		return nil, err
	}

	ts := types.TipSet{}
	err = json.Unmarshal(bs, &ts)
	if err != nil {
		return nil, err
	}

	cids := []string{}
	for _, b := range ts.Blocks() {
		cids = append(cids, fmt.Sprintf("%v", b.Cid()))
	}

	if len(cids) == 0 {
		return nil, xerrors.Errorf("Invalid block")
	}

	return cids, nil
}

func ChainBaseFee(host string) (float64, error) {
	bh, err := lotusbase.Request(lotusRpcUrl(host), []string{}, "Filecoin.ChainHead")
	if err != nil {
		return -1, err
	}

	head := types.TipSet{}
	json.Unmarshal(bh, &head)

	basefee := head.MinTicketBlock().ParentBaseFee
	feeStr := fmt.Sprintf("%v", basefee)
	ffee, _ := strconv.ParseFloat(feeStr, 64)
	ffee = ffee * 10e-10

	return ffee, nil
}

type ProvingDeadline struct {
	AllSectors       uint64
	FaultySectors    uint64
	Partitions       uint64
	ProvenPartitions uint64
	Current          bool
}

type Deadlines struct {
	Deadlines []ProvingDeadline
}

func ProvingDeadlines(host string, minerId string) (*Deadlines, error) {
	bh, err := lotusbase.Request(lotusRpcUrl(host), []interface{}{minerId, nil}, "Filecoin.StateMinerDeadlines")
	if err != nil {
		log.Errorf(log.Fields{}, "state miner deadlines fail: %v", err)
		return nil, err
	}

	deadlines := []api.Deadline{}
	json.Unmarshal(bh, &deadlines)

	bh, err = lotusbase.Request(lotusRpcUrl(host), []interface{}{minerId, nil}, "Filecoin.StateMinerProvingDeadline")
	if err != nil {
		log.Errorf(log.Fields{}, "state miner proving deadline fail: %v", err)
		return nil, err
	}

	di := dline.Info{}
	json.Unmarshal(bh, &di)

	provingDeadlines := Deadlines{
		Deadlines: []ProvingDeadline{},
	}

	for dlIdx, deadline := range deadlines {
		bh, err = lotusbase.Request(lotusRpcUrl(host), []interface{}{minerId, dlIdx, nil}, "Filecoin.StateMinerPartitions")
		if err != nil {
			log.Errorf(log.Fields{}, "state miner deadline partition fail: %v", err)
			return nil, err
		}

		partitions := []api.Partition{}
		json.Unmarshal(bh, &partitions)
		provenPartitions, err := deadline.PostSubmissions.Count()
		if err != nil {
			return nil, err
		}

		provingDeadline := ProvingDeadline{
			Partitions:       uint64(len(partitions)),
			ProvenPartitions: provenPartitions,
			Current:          uint64(dlIdx) == di.Index,
		}

		for _, partition := range partitions {
			sector, err := partition.AllSectors.Count()
			if err != nil {
				return nil, err
			}
			faulty, err := partition.FaultySectors.Count()
			if err != nil {
				return nil, err
			}

			provingDeadline.AllSectors += sector
			provingDeadline.FaultySectors += faulty
		}

		provingDeadlines.Deadlines = append(provingDeadlines.Deadlines, provingDeadline)
	}

	return &provingDeadlines, nil
}

func ImportWallet(host string, privateKey string, bearerToken string) (string, error) {
	data, err := hex.DecodeString(strings.TrimSpace(privateKey))
	if err != nil {
		return "", err
	}

	var ki types.KeyInfo
	if err := json.Unmarshal(data, &ki); err != nil {
		return "", err
	}

	addr, err := lotusbase.RequestWithBearerToken(lotusRpcUrl(host), []interface{}{ki}, "Filecoin.WalletImport", bearerToken)
	if err != nil {
		log.Errorf(log.Fields{}, "import wallet fail: %v", err)
		return "", err
	}

	return string(addr.([]byte)), err
}

func WalletExists(host string, address string, bearerToken string) (bool, error) {
	exists, err := lotusbase.RequestWithBearerToken(lotusRpcUrl(host), []interface{}{address}, "Filecoin.WalletHas", bearerToken)
	if err != nil {
		log.Errorf(log.Fields{}, "check wallet exists fail: %v", err)
		return false, err
	}

	return string(exists.([]byte)) == "true", err
}
