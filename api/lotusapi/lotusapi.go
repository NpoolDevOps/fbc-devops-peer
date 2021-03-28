package lotusapi

import (
	"encoding/json"
	"fmt"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusbase"
	"github.com/filecoin-project/lotus/api"
)

type SyncState struct {
	HeightDiff   int
	BlockElapsed int
}

func ChainSyncState(host string) (*SyncState, error) {
	b, err := lotusbase.Request(fmt.Sprintf("http://%v:1234/rpc/v0", host), []string{}, "Filecoin.SyncState")
	if err != nil {
		return nil, err
	}

	state := api.SyncState{}
	json.Unmarshal(b, &state)

	return &SyncState{}, nil
}
