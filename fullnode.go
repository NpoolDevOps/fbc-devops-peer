package main

type Fullnode struct {
}

func NewFullnodePeer(subRole string, parentSpec string) *Fullnode {
	return &Fullnode{}
}

func (n Fullnode) Run() error {
	return nil
}
