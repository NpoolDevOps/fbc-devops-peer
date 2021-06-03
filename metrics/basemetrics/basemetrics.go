package basemetrics

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	api "github.com/NpoolDevOps/fbc-devops-peer/api/minerapi"

	// "github.com/filecoin-project/go-address"
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

	NvmesTemprature    *prometheus.Desc
	RootIsWriteRead    *prometheus.Desc
	StorageIsWriteRead *prometheus.Desc

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
		NvmesTemprature: prometheus.NewDesc(
			"nvmes_temprature",
			"show every Nvme's temprature",
			[]string{"nvme", "tempname"}, nil,
		),
		RootIsWriteRead: prometheus.NewDesc(
			"root_is_write_read",
			"show whether root is write and read or not",
			nil, nil,
		),
		StorageIsWriteRead: prometheus.NewDesc(
			"storage_is_write_read",
			"show whether storage is write and read or not",
			[]string{"storage"}, nil,
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
	ch <- m.NvmesTemprature
	ch <- m.RootIsWriteRead
	ch <- m.StorageIsWriteRead
}

func (m *BaseMetrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(m.TimeDiff, prometheus.CounterValue, m.timeDiff)
	ch <- prometheus.MustNewConstMetric(m.PingGatewayDelay, prometheus.CounterValue, float64(m.pingGatewayDelayMs))
	ch <- prometheus.MustNewConstMetric(m.PingGatewayLost, prometheus.CounterValue, m.pingGatewayLost)
	ch <- prometheus.MustNewConstMetric(m.PingBaiduDelay, prometheus.CounterValue, float64(m.pingBaiduDelayMs))
	ch <- prometheus.MustNewConstMetric(m.PingBaiduLost, prometheus.CounterValue, m.pingBaiduLost)
	nvmeTempList, _ := getNvmeTempList()
	for nvme, tempList := range nvmeTempList {
		for tempname, temp := range tempList {
			tempFloat, _ := strconv.ParseFloat(temp, 64)
			ch <- prometheus.MustNewConstMetric(m.NvmesTemprature, prometheus.CounterValue, tempFloat, nvme, tempname)
		}
	}

	rootWR, _ := getFileIfWriteRead("/")
	var is float64
	if rootWR {
		is = 1
	} else {
		is = 0
	}
	ch <- prometheus.MustNewConstMetric(m.RootIsWriteRead, prometheus.CounterValue, is)

	storageAddressList, _ := getStorageAddress("/opt/sharestorage/")
	for _, address := range storageAddressList {
		storageWR, _ := getFileIfWriteRead(address)
		if storageWR {
			is = 1
		} else {
			is = 0
		}
		ch <- prometheus.MustNewConstMetric(m.StorageIsWriteRead, prometheus.CounterValue, is, address)
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

func getNvmeList() ([]string, error) {
	out, err := api.RunCommand(exec.Command("nvme", "list"))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to run nvme list"), err)
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	lines := 1
	nvmeList := []string{}
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if lines > 2 {
			nvmeList = append(nvmeList, strings.TrimSpace(strings.Split(string(line), "  ")[0]))
		}
		lines += 1
	}
	if lines == 1 {
		log.Errorf(log.Fields{}, "fail to get nvme list")
		return nil, err
	}
	return nvmeList, nil
}

func getTemp(nvme string) (map[string]string, error) {
	tempList := make(map[string]string)
	out, err := api.RunCommand(exec.Command("nvme", "smart-log", nvme))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to run nvme info"), err)
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "temperature") || strings.Contains(string(line), "Temperature ") && strings.Contains(string(line), " Temperature ") {
			name := strings.TrimSpace(strings.Split(string(line), ":")[0])
			temp := strings.TrimSpace(strings.Split(string(line), ":")[1])
			trueTemp := strings.TrimSpace(strings.Split(temp, " ")[0])
			tempList[name] = trueTemp
		}
	}
	log.Infof(log.Fields{}, "temp list is: %v", tempList)
	return tempList, nil
}

func getNvmeTempList() (map[string]map[string]string, error) {
	nvmeTempList := make(map[string]map[string]string)
	nvmeList, err := getNvmeList()
	tempList := make(map[string]string)
	if err != nil {
		return nil, err
	}
	for _, nvme := range nvmeList {
		tempList, err = getTemp(nvme)
		if err != nil {
			return nil, err
		}
		nvmeTempList[nvme] = tempList
	}
	return nvmeTempList, nil
}

func getFileIfWriteRead(file string) (bool, error) {
	out, err := api.RunCommand(exec.Command("getfacl", file))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get root zone"), err)
		return false, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(string(line), "user") && strings.Contains(string(line), "rw") {
			return true, nil
		}
	}
	return false, nil
}

func getStorageAddress(address string) ([]string, error) {
	storageAddressList := []string{}
	out, err := api.RunCommand(exec.Command("ls", address))
	if err != nil {
		log.Errorf(log.Fields{}, fmt.Sprintf("fail to get storage address"), err)
		return nil, err
	}
	br := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}
		lineArr := strings.Split(string(line), " ")
		for _, l := range lineArr {
			_, err = strconv.ParseInt(strings.TrimSpace(string(l)), 10, 64)
			if err == nil {
				storageAddressList = append(storageAddressList, address+strings.TrimSpace(string(l)))
			}
		}
	}
	return storageAddressList, nil
}
