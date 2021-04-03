package fbcsnmp

import (
	"fmt"
	log1 "github.com/EntropyPool/entropy-logger"
	g "github.com/gosnmp/gosnmp"
	"golang.org/x/xerrors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type SnmpConfig struct {
	target    string
	community string
	username  string
	password  string
	verbose   bool
}

type SnmpClient struct {
	config SnmpConfig
}

func NewSnmpClient(config SnmpConfig) *SnmpClient {
	return &SnmpClient{
		config: config,
	}
}

func (c *SnmpClient) CpuUsage() (int, int, int, error) {
	oids := []string{
		".1.3.6.1.4.1.2021.11.9.0",
		".1.3.6.1.4.1.2021.11.10.0",
		".1.3.6.1.4.1.2021.11.11.0",
	}

	outs, err := c.get(oids)
	if err != nil {
		return 100000, 100000, 100000, err
	}

	user, _ := strconv.ParseInt(outs[0], 10, 32)
	sys, _ := strconv.ParseInt(outs[1], 10, 32)
	idle, _ := strconv.ParseInt(outs[2], 10, 32)

	return int(user), int(sys), int(idle), nil
}

func (c *SnmpClient) get(oids []string) ([]string, error) {
	cli := &g.GoSNMP{
		Target:        c.config.target,
		Port:          161,
		Version:       g.Version2c,
		Community:     c.config.community,
		SecurityModel: g.UserSecurityModel,
		MsgFlags:      g.AuthPriv,
		Timeout:       time.Duration(30) * time.Second,
		SecurityParameters: &g.UsmSecurityParameters{
			UserName:                 c.config.username,
			AuthenticationProtocol:   g.SHA,
			AuthenticationPassphrase: c.config.password,
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        c.config.password,
		},
	}

	if c.config.verbose {
		cli.Logger = log.New(os.Stdout, "", 0)
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
		log1.Infof(log1.Fields{}, "%v: oid: %v", i, v.Name)
		switch v.Type {
		case g.OctetString:
			rcs = append(rcs, string(v.Value.([]byte)))
		default:
			rcs = append(rcs, fmt.Sprintf("%v", g.ToBigInt(v.Value)))
		}
	}

	return rcs, nil
}
