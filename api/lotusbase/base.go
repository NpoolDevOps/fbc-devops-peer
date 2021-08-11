package lotusbase

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
	"sync"
)

type RpcResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

type RpcParam struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Id      int         `json:"id"`
	Params  interface{} `json:"params"`
}

var curIdx = 0
var idMutex sync.Mutex

func curId() int {
	idMutex.Lock()
	defer idMutex.Unlock()
	id := curIdx
	curIdx += 1
	return id
}

func NewRpcParam(method string, params interface{}) *RpcParam {
	id := curId()

	return &RpcParam{
		Jsonrpc: "2.0",
		Method:  method,
		Id:      id,
		Params:  params,
	}
}

func Request(url string, params interface{}, method string) ([]byte, error) {
	ret, err := RequestWithBearerToken(url, params, method, "")
	return ret.([]byte), err
}

func RequestWithBearerToken(url string, params interface{}, method string, token string) (interface{}, error) {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetBody(NewRpcParam(method, params)).
		Post(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, xerrors.Errorf("fail to query chain sync status")
	}

	result := RpcResult{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	if result.Result == nil {
		return nil, xerrors.Errorf("%v to %v response nil", method, url)
	}

	b, err := json.Marshal(result.Result)
	if err != nil {
		log.Infof(log.Fields{}, "cannot marshal '%v', fallback", result.Result)
		return result.Result, nil
	}

	if b == nil {
		return nil, xerrors.Errorf("fail to marshal result %v to %v", method, url)
	}

	return b, nil
}
