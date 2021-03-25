package devopsruntime

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

func TestGetRootPart(t *testing.T) {
	rootPart := getRootPart()
	log.Infof(log.Fields{}, "root part: %v", rootPart)
}
