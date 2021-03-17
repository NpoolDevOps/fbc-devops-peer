package main

const (
	StorageMgrNode = "mgr"
	StorageMdsNode = "mds"
	StorageOsdNode = "osd"
)

type Storage struct {
}

func NewStoragePeer(config *PeerConfig) *Storage {
	return &Storage{}
}

func (n *Storage) Run() error {
	return nil
}
