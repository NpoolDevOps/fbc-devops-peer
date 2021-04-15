package logbase

import (
	"log"
	"testing"
)

func TestNewLogbase(t *testing.T) {
	newline := make(chan LogLine)
	lb := NewLogbase("/var/log/lotus/miner.log", newline)
	if lb == nil {
		log.Fatal("cannot watch file")
	}
	for {
		line, _ := <-newline
		log.Printf("%s\n", line.String())
	}
}
