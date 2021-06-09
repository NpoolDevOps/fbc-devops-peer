package types

import (
	"crypto/rsa"
	"hash"
)

type NotifyParentSpecInput struct {
	ParentSpec string `json:"parent_spec"`
}

type GetParentSpecOutput struct {
	ParentSpec string `json:"parent_spec"`
}

type OperationAction struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
}

type OperationInput struct {
	PublicKey *rsa.PublicKey  `json:"public_key"`
	Username  string          `json:"username"`
	Password  string          `json:"password"`
	Action    OperationAction `json:"action"`
	MsgHash   hash.Hash       `json:"msg_hash"`
}
