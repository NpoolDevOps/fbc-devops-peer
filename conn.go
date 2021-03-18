package main

import (
	"fmt"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
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
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	return GetParentSpecOutput{
		ParentSpec: spec.SN(),
	}, "", 0
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
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("http://%v%v", parentPeer, ParentSpecAPI))
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != 200 {
		return "", err
	}
	return string(resp.Body()), nil
}

func (p *PeerConnection) GetNotifiedParentSpec() (string, error) {
	return "", nil
}

func (p *PeerConnection) NotifyParentSpec(childPeer string) error {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(GetParentSpecInput{
			ParentSpec: spec.SN(),
		}).
		Post(fmt.Sprintf("http://%v%v", childPeer, ParentSpecAPI))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return xerrors.Errorf("http response error")
	}
	return nil
}
