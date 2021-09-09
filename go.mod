module github.com/NpoolDevOps/fbc-devops-peer

go 1.15

require (
	github.com/EntropyPool/entropy-logger v0.0.0-20210320022718-3091537e035f
	github.com/EntropyPool/machine-spec v0.0.0-20210408033101-2db09d058812
	github.com/NpoolDevOps/fbc-devops-service v0.0.0-20210413053955-6d57518351f4
	github.com/NpoolDevOps/fbc-license v0.0.0-20210408084724-dec53ab9682d
	github.com/NpoolRD/http-daemon v0.0.0-20210324100344-82fee56de8ac
	github.com/beevik/ntp v0.3.0
	github.com/coreos/bbolt v1.3.2 // indirect
	github.com/docker/go-units v0.4.0
	github.com/euank/go-kmsg-parser v2.0.0+incompatible
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/lotus v1.5.3
	github.com/go-ping/ping v0.0.0-20210327002015-80a511380375
	github.com/google/uuid v1.2.0
	github.com/gosnmp/gosnmp v1.30.0
	github.com/hpcloud/tail v1.0.0
	github.com/jaypipes/ghw v0.7.0
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/moby/sys/mountinfo v0.4.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/NpoolDevOps/fbc-license => ../fbc-license

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
