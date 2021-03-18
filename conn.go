package main

import (
	"github.com/NpoolRD/http-daemon"
	"net/http"
)

const peerHttpPort = 52375

type PeerConnection struct {
}

func NewPeerConnection() *PeerConnection {
	conn := &PeerConnection{}
	return conn
}

const (
	ParentSpecAPI = "/api/v0/peer/parentspec"
)

func (s *PeerConnection) ParentSpecGetRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}

func (s *PeerConnection) ParentSpecPostRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}

func (p *PeerConnection) Run() {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: ParentSpecAPI,
		Method:   "POST",
		Handler:  p.ParentSpecGetRequest,
	})
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: ParentSpecAPI,
		Method:   "GET",
		Handler:  p.ParentSpecPostRequest,
	})
	httpdaemon.Run(peerHttpPort)
}

func (p *PeerConnection) GetParentSpec(parentPeer string) (string, error) {
	return "", nil
}

func (p *PeerConnection) NotifyParentSpec(childPeer string, parentSpec string) error {
	return nil
}
