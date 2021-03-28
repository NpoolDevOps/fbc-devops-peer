package lotusapi

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusbase"
	"github.com/NpoolRD/http-daemon"
	"github.com/filecoin-project/lotus/api"
	"golang.org/x/xerrors"
)

type SyncState struct {
	HeightDiff   int
	BlockElapsed int
}

func ChainSyncState(host string) (*SyncState, error) {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(lotusbase.NewRpcParam("Filecoin.SyncState", []string{})).
		Post(fmt.Sprintf("http://%v:1234/rpc/v0", host))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, xerrors.Errorf("fail to query chain sync status")
	}

	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		return nil, xerrors.Errorf("invalid response from lotus")
	}

	if apiResp.Code != 0 {
		return nil, xerrors.Errorf("error response from lotus")
	}

	b, _ := json.Marshal(apiResp.Body)
	state := api.SyncState{}
	json.Unmarshal(b, &state)

	log.Infof(log.Fields{}, "CHAIN SYNC STATUS -- %v", state)

	return &SyncState{}, nil
}
