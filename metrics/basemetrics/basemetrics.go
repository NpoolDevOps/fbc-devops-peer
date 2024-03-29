package basemetrics

import (
	"bufio"
	"encoding/binary"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	"github.com/beevik/ntp"
	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/xerrors"
)

type BaseMetrics struct {
	TimeDiff         *prometheus.Desc
	PingGatewayDelay *prometheus.Desc
	PingGatewayLost  *prometheus.Desc
	PingBaiduDelay   *prometheus.Desc
	PingBaiduLost    *prometheus.Desc
	RootPermission   *prometheus.Desc
	RootMountRW      *prometheus.Desc

	NvmeTemperature *prometheus.Desc

	pingGatewayDelayMs int64
	pingBaiduDelayMs   int64
	pingGatewayLost    float64
	pingBaiduLost      float64

	username    string
	networkType string
}

func NewBaseMetrics(username, networkType string) *BaseMetrics {
	metrics := &BaseMetrics{
		username:    username,
		networkType: networkType,
		PingGatewayDelay: prometheus.NewDesc(
			"base_ping_gateway_delay",
			"Show base ping gateway delay",
			[]string{"networktype", "user"}, nil,
		),
		PingGatewayLost: prometheus.NewDesc(
			"base_ping_gateway_lost",
			"Show base ping gateway lost",
			[]string{"networktype", "user"}, nil,
		),
		PingBaiduDelay: prometheus.NewDesc(
			"base_ping_baidu_delay",
			"Show base ping baidu lost",
			[]string{"networktype", "user"}, nil,
		),
		PingBaiduLost: prometheus.NewDesc(
			"base_ping_baidu_lost",
			"Show base ping baidu lost",
			[]string{"networktype", "user"}, nil,
		),
		TimeDiff: prometheus.NewDesc(
			"base_ntp_time_diff",
			"Show base ntp time diff",
			[]string{"networktype", "user"}, nil,
		),
		RootPermission: prometheus.NewDesc(
			"base_root_permission",
			"show whether the root is able to write and read",
			[]string{"networktype", "user"}, nil,
		),
		RootMountRW: prometheus.NewDesc(
			"base_root_mount_rw",
			"show whether root mount access is rw",
			[]string{"networktype", "user"}, nil,
		),
		NvmeTemperature: prometheus.NewDesc(
			"base_nvme_temperature",
			"show nvme temperature",
			[]string{"nvme", "networktype", "user"}, nil,
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
	ch <- m.RootPermission
	ch <- m.RootMountRW
	ch <- m.NvmeTemperature
}

func (m *BaseMetrics) Collect(ch chan<- prometheus.Metric) {
	username := m.username
	networkType := m.networkType

	timeDiff, _ := getNtpDiff()
	ch <- prometheus.MustNewConstMetric(m.TimeDiff, prometheus.CounterValue, timeDiff, networkType, username)
	ch <- prometheus.MustNewConstMetric(m.PingGatewayDelay, prometheus.CounterValue, float64(m.pingGatewayDelayMs), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.PingGatewayLost, prometheus.CounterValue, m.pingGatewayLost, networkType, username)
	ch <- prometheus.MustNewConstMetric(m.PingBaiduDelay, prometheus.CounterValue, float64(m.pingBaiduDelayMs), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.PingBaiduLost, prometheus.CounterValue, m.pingBaiduLost, networkType, username)
	rootPerm, _ := systemapi.FilePerm2Int("/")
	ch <- prometheus.MustNewConstMetric(m.RootPermission, prometheus.CounterValue, float64(rootPerm), networkType, username)

	mountpointWrittable, _ := systemapi.MountpointWrittable("/")
	if mountpointWrittable {
		ch <- prometheus.MustNewConstMetric(m.RootMountRW, prometheus.CounterValue, 1, networkType, username)
	} else {
		ch <- prometheus.MustNewConstMetric(m.RootMountRW, prometheus.CounterValue, 0, networkType, username)
	}

	nvmeTemperatureList, _ := systemapi.GetNvmeTemperatureList()
	for nvmeName, temperature := range nvmeTemperatureList {
		ch <- prometheus.MustNewConstMetric(m.NvmeTemperature, prometheus.CounterValue, temperature, nvmeName, networkType, username)
	}
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

var NtpServers = []string{"asia.pool.ntp.org", "cn.pool.ntp.org", "ae.pool.ntp.org", "in.pool.ntp.org", "sa.pool.ntp.org"}

func getNtpDiff() (float64, error) {
	done := make(chan time.Time)

	for _, server := range NtpServers {
		go func(server string) {
			ntpTime, err := ntp.Time(server)
			if err == nil {
				done <- ntpTime
			}
		}(server)
	}

	select {
	case ntpTime := <-done:
		nowTimeMs := time.Now().Local().UnixNano() / int64(time.Millisecond)
		ntpTimeMs := ntpTime.Local().UnixNano() / int64(time.Millisecond)
		timeDiff := math.Abs(float64(ntpTimeMs - nowTimeMs))
		return timeDiff, nil
	case <-time.After(2 * time.Second):
		return -1, xerrors.Errorf("get ntp time beyond 2 seconds")
	}

}
