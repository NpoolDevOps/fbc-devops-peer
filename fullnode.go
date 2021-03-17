package main

type Fullnode struct {
}

func NewFullnodePeer(config *PeerConfig) *Fullnode {
	return &Fullnode{}
}

func (n Fullnode) Run() error {
	return nil
}
