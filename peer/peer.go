package peer

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/NpoolDevOps/fbc-devops-peer/node"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	"github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
)

const peerHttpPort = 52375

type Peer struct {
	Node               node.Node
	NotifiedParentSpec string
}

func NewPeer(node node.Node) *Peer {
	conn := &Peer{
		Node: node,
	}
	return conn
}

func (p *Peer) ParentSpecGetRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	return types.GetParentSpecOutput{
		ParentSpec: spec.SN(),
	}, "", 0
}

func (p *Peer) ParentSpecPostRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}
	input := types.NotifyParentSpecInput{}
	json.Unmarshal(b, &input)
	p.Node.NotifyParentSpec(input.ParentSpec)
	return nil, "", 0
}

func (p *Peer) Run() {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.ParentSpecAPI,
		Method:   "POST",
		Handler:  p.ParentSpecPostRequest,
	})
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.ParentSpecAPI,
		Method:   "GET",
		Handler:  p.ParentSpecGetRequest,
	})
	httpdaemon.Run(peerHttpPort)
}

func (p *Peer) GetParentSpec(parentPeer string) (string, error) {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("http://%v%v", parentPeer, types.ParentSpecAPI))
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != 200 {
		return "", err
	}
	return string(resp.Body()), nil
}

func (p *Peer) NotifyParentSpec(childPeer string) error {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(types.NotifyParentSpecInput{
			ParentSpec: spec.SN(),
		}).
		Post(fmt.Sprintf("http://%v%v", childPeer, types.ParentSpecAPI))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return xerrors.Errorf("http response error")
	}
	return nil
}
