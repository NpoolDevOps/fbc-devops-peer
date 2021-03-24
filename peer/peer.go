package peer

import (
	"encoding/json"
	"fmt"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
)

const peerHttpPort = 52375

type PeerConnection struct {
	NotifiedParentSpec string
}

func NewPeerConnection() *PeerConnection {
	conn := &PeerConnection{}
	return conn
}

const (
	ParentSpecAPI = "/api/v0/peer/parentspec"
)

func (p *PeerConnection) ParentSpecGetRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	return GetParentSpecOutput{
		ParentSpec: spec.SN(),
	}, "", 0
}

func (p *PeerConnection) ParentSpecPostRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}
	input := NotifyParentSpecInput{}
	json.Unmarshal(b, &input)
	p.NotifiedParentSpec = input.ParentSpec
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
	if p.NotifiedParentSpec == "" {
		return "", xerrors.Errorf("invalid parent spec")
	}
	return p.NotifiedParentSpec, nil
}

func (p *PeerConnection) NotifyParentSpec(childPeer string) error {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(NotifyParentSpecInput{
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
