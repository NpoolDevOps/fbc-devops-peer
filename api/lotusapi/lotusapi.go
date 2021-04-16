package lotusapi

import (
	"encoding/json"
	"fmt"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusbase"
	"github.com/NpoolDevOps/fbc-devops-peer/version"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/xerrors"
	"time"
)

type SyncState struct {
	HeightDiff   int64
	BlockElapsed time.Duration
	SyncError    bool
	NetPeers     int
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
	json.Unmarshal(bs, &ts)

	cids := []string{}
	for _, b := range ts.Blocks() {
		cids = append(cids, fmt.Sprintf("%v", b.Cid()))
	}

	return cids, nil
}
