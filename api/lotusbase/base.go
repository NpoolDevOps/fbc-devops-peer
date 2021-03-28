package lotusbase

import (
	"sync"
)

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
