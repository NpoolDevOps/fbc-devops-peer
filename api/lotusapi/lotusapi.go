package lotusapi

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
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

	log.Infof(log.Fields{}, "RESP BODY --- %v", string(b))
	log.Infof(log.Fields{}, "CHAIN SYNC STATE --- %v", state)

	return &SyncState{}, nil
}
