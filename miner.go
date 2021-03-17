package main

type Miner struct {
}

func NewMinerPeer(subRole string, parentSpec string) *Miner {
	return &Miner{}
}

func (n *Miner) Run() error {
	return nil
}
