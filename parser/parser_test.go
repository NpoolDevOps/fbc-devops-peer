package parser

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

func TestParseAPIInfo(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("cannot create parser")
	}
}

func TestReadEnvFromAPIFile(t *testing.T) {
	parser := NewParser()
	err := parser.readEnvFromAPIFile(FullnodeAPIFile)
	if err != nil {
		t.Errorf("cannot read env from %v", FullnodeAPIFile)
	}
	err = parser.readEnvFromAPIFile(MinerAPIFile)
	if err != nil {
		t.Errorf("cannot read env from %v", MinerAPIFile)
	}
}

func TestParseIPFromEnvValue(t *testing.T) {
	parser := NewParser()
	ip, err := parser.parseIPFromEnvValue("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.Ai5dIx4tHggODtmGpGx9-uOvnmcB4r62fljRPYmR4TE:/ip4/10.133.14.57/tcp/2345/http")
	if err != nil {
		t.Errorf("cannot parse ip from env")
	}
	log.Infof(log.Fields{}, "ip = %v", ip)
}

func TestParse(t *testing.T) {
	NewParser()
}
