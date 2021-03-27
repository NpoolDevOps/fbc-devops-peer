package node

import (
	"github.com/google/uuid"
)

type Node interface {
	GetMainRole() string
	GetSubRole() string
	NotifyParentSpec(string)
	GetParentIP() (string, error)
	GetChildsIPs() ([]string, error)
	NotifyPeerId(uuid.UUID)
	Banner()
}
