package main

type Storage struct {
}

func NewStoragePeer(subRole string, parentSpec string) *Storage {
	return &Storage{}
}

func (n *Storage) Run() error {
	return nil
}
