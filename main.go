package main

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"os"
)

func main() {
	app := &cli.App{
		Name:                 "fbc-devops-peer",
		Usage:                "FBC devops peer used to report peer information",
		Version:              "0.1.0",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "main-role",
				Usage: "First level role in cluster [fullnode | miner | worker | storage]",
			},
			&cli.StringFlag{
				Name:  "sub-role",
				Usage: "Sub level role in cluster, [mds | mgr | osd] for storage, ignored by others",
			},
			&cli.StringFlag{
				Name:  "parent-spec",
				Usage: "Hardware spec of parent node",
			},
			&cli.IntFlag{
				Name: "nvme-count",
			},
			&cli.IntFlag{
				Name: "gpu-count",
			},
		},
		Action: func(cctx *cli.Context) error {
			config := &PeerConfig{
				MainRole:   cctx.String("main-role"),
				SubRole:    cctx.String("sub-role"),
				ParentSpec: cctx.String("parent-spec"),
			}

			config.NvmeCount = cctx.Int("nvme-count")
			config.GpuCount = cctx.Int("gpu-count")

			switch config.MainRole {
			case FullNode:
				return NewFullnodePeer(config).Run()
			case MinerNode:
				return NewMinerPeer(config).Run()
			case WorkerNode:
				return NewWorkerPeer(config).Run()
			case StorageNode:
				return NewStoragePeer(config).Run()
			default:
				return xerrors.Errorf("Unknow main role %v", config.MainRole)
			}
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf(log.Fields{}, "fail to run %v: %v", app.Name, err)
	}
}
