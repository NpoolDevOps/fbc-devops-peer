package main

const (
	FullnodeWallet     = "wallet"
	FullnodeAccounting = "accounting"
)

type Fullnode struct {
}

func NewFullnodePeer(config *BasenodeConfig) *Fullnode {
	return &Fullnode{}
}

func (n Fullnode) Run() error {
	return nil
}
