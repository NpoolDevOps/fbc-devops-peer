package peer

import (
	"bufio"
	"crypto"
	"crypto/rsa"

	// "crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/NpoolDevOps/fbc-devops-peer/node"
	"github.com/NpoolDevOps/fbc-devops-peer/operation"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"github.com/amoghe/go-crypt"
	"golang.org/x/xerrors"
)

const peerHttpPort = 52375

type Peer struct {
	Node             node.Node
	parentSpecTicker *time.Ticker
	spec             string
	operation        *operation.Operation
}

func NewPeer(node node.Node) *Peer {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()

	conn := &Peer{
		Node:             node,
		parentSpecTicker: time.NewTicker(2 * time.Minute),
		spec:             spec.SN(),
		operation:        operation.NewOperation(),
	}

	return conn
}

func (p *Peer) handler() {
	for {
		ip, err := p.Node.GetParentIP()
		if err == nil {
			spec, err := p.GetParentSpec(ip)
			if err == nil {
				p.Node.NotifyParentSpec(spec)
			} else {
				log.Infof(log.Fields{}, "cannot get parent spec from %v: %v", ip, err)
			}
		}
		childs, err := p.Node.GetChildsIPs()
		if err == nil {
			for _, child := range childs {
				err = p.NotifyParentSpec(child)
				if err != nil {
					log.Errorf(log.Fields{}, "fail to notify parent spec to %v: %v", child, err)
				}
			}
		} else {
			log.Infof(log.Fields{}, "cannot get childs for %v : %v", p.Node.GetMainRole(), err)
		}
		<-p.parentSpecTicker.C
	}
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

func (p *Peer) HeartbeatRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}
func (p *Peer) verifyUser(username, password string) (bool, error) {
	filepath := "/etc/shadow"
	file, err := os.OpenFile(filepath, os.O_RDWR, 0666)
	var condi, myPass string
	if err != nil {
		return false, err
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		lineArr := strings.Split(line, ":")
		if lineArr[0] == username {
			la := strings.Split(lineArr[1], "$")
			condi = "$" + strings.TrimSpace(la[1]) + "$" + strings.TrimSpace(la[2])
			myPass = lineArr[1]
			fmt.Println(condi, myPass)
		}
	}
	sha512, err := crypt.Crypt(password, condi)
	if err != nil {
		return false, err
	}
	fmt.Println(sha512, myPass)
	if sha512 == myPass {
		fmt.Println("right")
		return true, nil
	} else {
		fmt.Println("wrong")
		return false, nil
	}
}

func (p *Peer) OperationRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}
	input := types.OperationInput{}
	json.Unmarshal(b, &input)
	// TODO: verify public key or username / password
	//to verify public key
	// signature := input.Action.Params			//get the sign from the last 256byte of the data in params, and is []byte value
	signature := []byte{}                //signature which after Encrypting
	msghashSum := input.MsgHash.Sum(nil) //hash sum the signature string
	err = rsa.VerifyPSS(input.PublicKey, crypto.SHA256, msghashSum, signature, nil)
	if err != nil {
		log.Errorf(log.Fields{}, "this public key can not decrypt the signature", err)
		_, errUser := p.verifyUser(input.Username, input.Password)
		if errUser != nil {
			log.Errorf(log.Fields{}, "verify user identify error...", errUser, err)
			return nil, errUser.Error() + err.Error(), -2
		}
	}

	resp, err := p.operation.Exec(input.Action)
	if err != nil {
		return nil, err.Error(), -3
	}

	return resp, "", 0
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
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.HeartbeatAPI,
		Method:   "GET",
		Handler:  p.HeartbeatRequest,
	})
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.OperationAPI,
		Method:   "POST",
		Handler:  p.OperationRequest,
	})
	httpdaemon.Run(peerHttpPort)
	go p.handler()
}

func (p *Peer) GetParentSpec(parentPeer string) (string, error) {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("http://%v:%v%v", parentPeer, peerHttpPort, types.ParentSpecAPI))
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get parent spec: %v", err)
		return "", err
	}
	if resp.StatusCode() != 200 {
		log.Errorf(log.Fields{}, "fail to get parent spec: NON-200: %v", resp.StatusCode())
		return "", err
	}
	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get parent spec: %v", err)
		return "", err
	}

	if apiResp.Code != 0 {
		log.Errorf(log.Fields{}, "fail to get parent spec: %v", apiResp.Msg)
		return "", xerrors.Errorf("fail to get parent spec: %v", apiResp.Msg)
	}

	output := types.GetParentSpecOutput{}
	b, _ := json.Marshal(apiResp.Body)
	json.Unmarshal(b, &output)

	return output.ParentSpec, nil
}

func (p *Peer) NotifyParentSpec(childPeer string) error {
	spec := machspec.NewMachineSpec()
	spec.PrepareLowLevel()
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(types.NotifyParentSpecInput{
			ParentSpec: spec.SN(),
		}).
		Post(fmt.Sprintf("http://%v:%v%v", childPeer, peerHttpPort, types.ParentSpecAPI))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return xerrors.Errorf("http response error")
	}
	return nil
}

func (p *Peer) Heartbeat(childPeer string) error {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("http://%v:%v%v", childPeer, peerHttpPort, types.HeartbeatAPI))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return xerrors.Errorf("http response error")
	}
	return nil
}
