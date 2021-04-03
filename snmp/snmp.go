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
	Target          string
	Community       string
	Username        string
	Password        string
	verbose         bool
	ConfigBandwidth int64
}

type SnmpClient struct {
	config *SnmpConfig
	client *g.GoSNMP
}

func NewSnmpClient(config *SnmpConfig) *SnmpClient {
	cli := &g.GoSNMP{
		Target:        config.Target,
		Port:          161,
		Version:       g.Version2c,
		Community:     config.Community,
		SecurityModel: g.UserSecurityModel,
		MsgFlags:      g.AuthPriv,
		Timeout:       time.Duration(30) * time.Second,
		SecurityParameters: &g.UsmSecurityParameters{
			UserName:                 config.Username,
			AuthenticationProtocol:   g.SHA,
			AuthenticationPassphrase: config.Password,
			PrivacyProtocol:          g.DES,
			PrivacyPassphrase:        config.Password,
		},
	}

	if config.verbose {
		cli.Logger = log.New(os.Stdout, "", 0)
	}

	return &SnmpClient{
		config: config,
		client: cli,
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

func (c *SnmpClient) NetworkBandwidth() (int64, int64, error) {
	oid := ".1.3.6.1.2.1.2.2.1.5"
	bwStr, err := c.walk(oid)
	if err != nil {
		return 0, 0, err
	}

	bw, _ := strconv.ParseInt(bwStr, 10, 32)
	return bw, c.config.ConfigBandwidth, nil
}

func (c *SnmpClient) NetworkBytes() (int64, int64, error) {
	oid := ".1.3.6.1.2.1.2.2.1.10"
	bwStr, err := c.walk(oid)
	if err != nil {
		return 0, 0, err
	}
	recv, _ := strconv.ParseInt(bwStr, 10, 32)

	oid = ".1.3.6.1.2.1.2.2.1.16"
	bwStr, err = c.walk(oid)
	if err != nil {
		return 0, 0, err
	}
	send, _ := strconv.ParseInt(bwStr, 10, 32)

	return recv, send, nil
}

func (c *SnmpClient) parsePdu(pdu g.SnmpPDU) string {
	switch pdu.Type {
	case g.OctetString:
		return string(pdu.Value.([]byte))
	default:
		return fmt.Sprintf("%v", g.ToBigInt(pdu.Value))
	}
}

func (c *SnmpClient) parsePacket(pkt *g.SnmpPacket) ([]string, error) {
	rcs := []string{}

	for i, v := range pkt.Variables {
		if strings.HasSuffix(v.Name, "1.3.6.1.6.3.15.1.1.3.0") {
			return nil, xerrors.Errorf("unknow username or password")
		}
		log1.Infof(log1.Fields{}, "%v: oid: %v", i, v.Name)
		rcs = append(rcs, c.parsePdu(v))
	}

	return rcs, nil
}

func (c *SnmpClient) walk(oid string) (string, error) {
	cli := c.client
	if err := cli.Connect(); err != nil {
		return "", err
	}
	defer cli.Conn.Close()

	rc := ""

	err := cli.BulkWalk(oid, func(pdu g.SnmpPDU) error {
		rc = c.parsePdu(pdu)
		return nil
	})

	return rc, err
}

func (c *SnmpClient) get(oids []string) ([]string, error) {
	cli := c.client
	if err := cli.Connect(); err != nil {
		return nil, err
	}
	defer cli.Conn.Close()

	pkt, err := cli.Get(oids)
	if err != nil {
		return nil, err
	}

	return c.parsePacket(pkt)
}
