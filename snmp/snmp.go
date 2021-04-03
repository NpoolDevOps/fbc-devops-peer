package fbcsnmp

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	g "github.com/gosnmp/gosnmp"
	"golang.org/x/xerrors"
	"strings"
	"time"
)

type SnmpClient struct {
	target    string
	community string
}

func NewSnmpClient(target string, community string) *SnmpClient {
	cli := &SnmpClient{
		target:    target,
		community: community,
	}

	return cli
}

func (c *SnmpClient) CpuUsage() ([]string, error) {
	oids := []string{
		".1.3.6.1.4.1.2021.11.9.0",
		".1.3.6.1.4.1.2021.11.10.0",
		".1.3.6.1.4.1.2021.11.11.0",
	}

	outs, err := c.get(oids)
	if err != nil {
		return outs, err
	}

	return outs, nil
}

func (c *SnmpClient) get(oids []string) ([]string, error) {
	cli := &g.GoSNMP{
		Target:        c.target,
		Port:          161,
		Version:       g.Version3,
		SecurityModel: g.UserSecurityModel,
		MsgFlags:      g.AuthPriv,
		Timeout:       time.Duration(30) * time.Second,
		SecurityParameters: &g.UsmSecurityParameters{
			UserName:                 "user",
			AuthenticationProtocol:   g.SHA,
			AuthenticationPassphrase: "password",
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        "password",
		},
	}

	if err := cli.Connect(); err != nil {
		return nil, err
	}
	defer cli.Conn.Close()

	rc, err := cli.Get(oids)
	if err != nil {
		return nil, err
	}

	rcs := []string{}

	for i, v := range rc.Variables {
		if strings.HasSuffix(v.Name, "1.3.6.1.6.3.15.1.1.3.0") {
			return nil, xerrors.Errorf("unknow username or password")
		}
		log.Infof(log.Fields{}, "%v: oid: %v", i, v.Name)
		switch v.Type {
		case g.OctetString:
			rcs = append(rcs, string(v.Value.([]byte)))
		default:
			rcs = append(rcs, fmt.Sprintf("%v", g.ToBigInt(v.Value)))
		}
	}

	return rcs, nil
}
