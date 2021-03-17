package main

const (
	StorageMgrNode = "mgr"
	StorageMdsNode = "mds"
	StorageOsdNode = "osd"
)

const (
	StorageVendorUcloud   = "ucloud"
	StorageVendorQiniu    = "qiniu"
	StorageVendorLangchao = "langchao"
	StorageVendorShuguang = "shuguang"
)

type Storage struct {
}

func NewStoragePeer(config *BasenodeConfig) *Storage {
	return &Storage{}
}

func (n *Storage) Run() error {
	return nil
}
