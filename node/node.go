package node

type Node interface {
	GetMainRole() string
	GetSubRole() string
	NotifyParentSpec(string)
	GetParentIP() (string, error)
	GetChildsIPs() ([]string, error)
}
