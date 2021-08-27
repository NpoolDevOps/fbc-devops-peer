package main

import (
	"fmt"

	log "github.com/EntropyPool/entropy-logger"
)

// var NtpServers = []string{"asia.pool.ntp.org", "cn.pool.ntp.org", "ae.pool.ntp.org", "in.pool.ntp.org", "sa.pool.ntp.org"}

// func main() {
// 	ntpServer := ""
// 	for _, server := range NtpServers {
// 		ntpTime, err := ntp.Time(ntpServer)
// 		if err != nil {
// 			log.Errorf(log.Fields{}, "get ntp time error")
// 		}
// 		fmt.Println("server is", server, "ntp time is", ntpTime.Unix())
// 	}
// }
type DeviceIp struct {
	GigabitIp    string
	TenGigabitIp string
}

func main() {
	var localAddr string
	deviceIp := DeviceIp{
		GigabitIp:    "",
		TenGigabitIp: "",
	}

	switch {
	case deviceIp.GigabitIp != "":
		localAddr = deviceIp.GigabitIp
	case deviceIp.TenGigabitIp != "":
		localAddr = deviceIp.TenGigabitIp
	default:
		log.Errorf(log.Fields{}, "lost local address")
	}
	fmt.Println("localaddr is", localAddr)
}
