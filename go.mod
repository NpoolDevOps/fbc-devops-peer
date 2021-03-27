module github.com/NpoolDevOps/fbc-devops-peer

go 1.15

require (
	github.com/EntropyPool/entropy-logger v0.0.0-20210320022718-3091537e035f
	github.com/EntropyPool/machine-spec v0.0.0-20210326115744-eaf9c297f65c
	github.com/NpoolDevOps/fbc-devops-service v0.0.0-20210327050328-f968ca812d64
	github.com/NpoolDevOps/fbc-license v0.0.0
	github.com/NpoolDevOps/fbc-license-service v0.0.0-20210327043102-a4a5cab50cf4
	github.com/NpoolRD/http-daemon v0.0.0-20210324100344-82fee56de8ac
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/google/uuid v1.2.0
	github.com/jaypipes/ghw v0.7.0
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/rai-project/config v0.0.0-20190926180509-3bd01e698aad // indirect
	github.com/rai-project/logger v0.0.0-20190701163301-49978a80bf96 // indirect
	github.com/rai-project/nvidia-smi v0.0.0-20190730061239-864eb441c9ae
	github.com/urfave/cli/v2 v2.3.0
	github.com/xjh22222228/ip v1.0.1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/NpoolDevOps/fbc-license => ../fbc-license

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
