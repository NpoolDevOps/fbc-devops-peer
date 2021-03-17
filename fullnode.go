package main

type Fullnode struct {
}

func NewFullnodePeer(config *BasenodeConfig) *Fullnode {
	return &Fullnode{}
}

func (n Fullnode) Run() error {
	return nil
}
