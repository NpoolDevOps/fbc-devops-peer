module github.com/NpoolDevOps/fbc-devops-peer

go 1.15

require (
	github.com/EntropyPool/entropy-logger v0.0.0-20210320022718-3091537e035f
	github.com/EntropyPool/machine-spec v0.0.0-20210408033101-2db09d058812
	github.com/NpoolDevOps/fbc-devops-service v0.0.0-20210413053955-6d57518351f4
	github.com/NpoolDevOps/fbc-license v0.0.0-20210408164724-dec53ab9682d
	github.com/NpoolDevOps/fbc-license-service v0.0.0-20210328062839-d1527bc31f7e
	github.com/NpoolRD/http-daemon v0.0.0-20210324100344-82fee56de8ac
	github.com/docker/go-units v0.4.0
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/lotus v1.5.3
	github.com/go-ping/ping v0.0.0-20210327002015-80a511380375
	github.com/google/uuid v1.2.0
	github.com/gosnmp/gosnmp v1.30.0
	github.com/hpcloud/tail v1.0.0
	github.com/ipfs/go-cid v0.0.7
	github.com/jaypipes/ghw v0.7.0
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/rai-project/config v0.0.0-20190926180509-3bd01e698aad // indirect
	github.com/rai-project/logger v0.0.0-20190701163301-49978a80bf96 // indirect
	github.com/rai-project/nvidia-smi v0.0.0-20190730061239-864eb441c9ae
	github.com/rai-project/tegra v0.0.0-20181119122707-1d9901ca382b // indirect
	github.com/urfave/cli/v2 v2.3.0
	github.com/xjh22222228/ip v1.0.1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/NpoolDevOps/fbc-license => ../fbc-license

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
