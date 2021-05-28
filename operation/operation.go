package operation

import (
	"encoding/json"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	"golang.org/x/xerrors"
)

type Operation struct {
}

func NewOperation() *Operation {
	op := &Operation{}
	return op
}

const (
	ActionAcceptance = "acceptance"
	ActionPreinstall = "preinstall"
	ActionInstall    = "install"
	ActionInstallBin = "installbin"
	ActionReleaseBin = "releasebin"
)

func (op *Operation) onAcceptance(params string) (interface{}, error) {
	return acceptanceExec(params)
}

func (op *Operation) Exec(action types.OperationAction) (interface{}, error) {
	b, _ := json.Marshal(action.Params)

	switch action.Action {
	case ActionAcceptance:
		return op.onAcceptance(string(b))
	case ActionPreinstall:
	case ActionInstall:
	case ActionInstallBin:
	case ActionReleaseBin:
	default:
		return nil, xerrors.Errorf("unknow action %v", action.Action)
	}

	return nil, xerrors.Errorf("unknow action %v", action.Action)
}
