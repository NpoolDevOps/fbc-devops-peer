package fbcsnmp

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	g "github.com/gosnmp/gosnmp"
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
		"1.3.6.1.4.1.2021.11.9.0",
		"1.3.6.1.4.1.2021.11.10.0",
		"1.3.6.1.4.1.2021.11.11.0",
	}

	outs, err := c.get(oids)
	if err != nil {
		return outs, err
	}

	return outs, nil
}

func (c *SnmpClient) get(oids []string) ([]string, error) {
	g.Default.Target = c.target

	if err := g.Default.Connect(); err != nil {
		return nil, err
	}
	defer g.Default.Conn.Close()

	rc, err := g.Default.Get(oids)
	if err != nil {
		return nil, err
	}

	rcs := []string{}

	for i, v := range rc.Variables {
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
