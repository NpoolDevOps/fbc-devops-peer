package basemetrics

import (
	"bufio"
	"encoding/binary"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/xerrors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type BaseMetrics struct {
	TimeDiff         *prometheus.Desc
	PingGatewayDelay *prometheus.Desc
	PingGatewayLost  *prometheus.Desc
	PingBaiduDelay   *prometheus.Desc
	PingBaiduLost    *prometheus.Desc

	pingGatewayDelayMs int64
	pingBaiduDelayMs   int64
	pingGatewayLost    float64
	pingBaiduLost      float64
	timeDiff           float64
}

func NewBaseMetrics() *BaseMetrics {
	metrics := &BaseMetrics{
		PingGatewayDelay: prometheus.NewDesc(
			"base_ping_gateway_delay",
			"Show base ping gateway delay",
			nil, nil,
		),
		PingGatewayLost: prometheus.NewDesc(
			"base_ping_gateway_lost",
			"Show base ping gateway lost",
			nil, nil,
		),
		PingBaiduDelay: prometheus.NewDesc(
			"base_ping_baidu_delay",
			"Show base ping baidu lost",
			nil, nil,
		),
		PingBaiduLost: prometheus.NewDesc(
			"base_ping_baidu_lost",
			"Show base ping baidu lost",
			nil, nil,
		),
		TimeDiff: prometheus.NewDesc(
			"base_ntp_time_diff",
			"Show base ntp time diff",
			nil, nil,
		),
	}

	go metrics.updater()

	return metrics
}

func (m *BaseMetrics) updater() {
	ticker := time.NewTicker(2 * time.Minute)
	for {
		ip, err := getDefaultGateway()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to get default gateway: %v", err)
			<-ticker.C
			continue
		}

		delay, lost := pingStatistic(ip)
		m.pingGatewayDelayMs = delay
		m.pingGatewayLost = lost

		delay, lost = pingStatistic("www.baidu.com")
		m.pingBaiduDelayMs = delay
		m.pingBaiduLost = lost

		<-ticker.C
	}
}

func (m *BaseMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.TimeDiff
	ch <- m.PingGatewayDelay
	ch <- m.PingGatewayLost
	ch <- m.PingBaiduDelay
	ch <- m.PingBaiduLost
}

func (m *BaseMetrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(m.TimeDiff, prometheus.CounterValue, m.timeDiff)
	ch <- prometheus.MustNewConstMetric(m.PingGatewayDelay, prometheus.CounterValue, float64(m.pingGatewayDelayMs))
	ch <- prometheus.MustNewConstMetric(m.PingGatewayLost, prometheus.CounterValue, m.pingGatewayLost)
	ch <- prometheus.MustNewConstMetric(m.PingBaiduDelay, prometheus.CounterValue, float64(m.pingBaiduDelayMs))
	ch <- prometheus.MustNewConstMetric(m.PingBaiduLost, prometheus.CounterValue, m.pingBaiduLost)
}

func pingStatistic(host string) (ms int64, rate float64) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to create pinger %v: %v", host, err)
		return -1, -1
	}

	pinger.SetPrivileged(true)
	pinger.Count = 200

	err = pinger.Run()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to run pinger %v: %v", host, err)
		return -2, -2
	}

	stats := pinger.Statistics()
	return stats.AvgRtt.Milliseconds(), stats.PacketLoss
}

const (
	file  = "/proc/net/route"
	line  = 1    // line containing the gateway addr. (first line: 0)
	sep   = "\t" // field separator
	field = 2    // field containing hex gateway address (first field: 0)
)

func getDefaultGateway() (string, error) {

	file, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		// jump to line containing the agteway address
		for i := 0; i < line; i++ {
			scanner.Scan()
		}

		// get field containing gateway address
		tokens := strings.Split(scanner.Text(), sep)
		gatewayHex := "0x" + tokens[field]

		// cast hex address to uint32
		d, _ := strconv.ParseInt(gatewayHex, 0, 64)
		d32 := uint32(d)

		// make net.IP address from uint32
		ipd32 := make(net.IP, 4)
		binary.LittleEndian.PutUint32(ipd32, d32)

		// format net.IP to dotted ipV4 string
		ip := net.IP(ipd32).String()

		// exit scanner
		return ip, nil
	}

	return "", xerrors.Errorf("fail to read gateway")
}
