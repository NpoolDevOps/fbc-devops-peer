package main

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	"github.com/NpoolDevOps/fbc-devops-peer/peer"
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
			&cli.StringFlag{
				Name: "report-host",
			},
			&cli.IntFlag{
				Name: "nvme-count",
			},
			&cli.IntFlag{
				Name: "gpu-count",
			},
			&cli.StringFlag{
				Name: "device-user",
			},
		},
		Action: func(cctx *cli.Context) error {
			config := &basenode.BasenodeConfig{
				NodeConfig: &basenode.NodeConfig{
					MainRole:   cctx.String("main-role"),
					SubRole:    cctx.String("sub-role"),
					ParentSpec: cctx.String("parent-spec"),
					HardwareConfig: &basenode.NodeHardware{
						NvmeCount: cctx.Int("nvme-count"),
						GpuCount:  cctx.Int("gpu-count"),
					},
				},
				User: cctx.String("device-user"),
			}

			client := devops.NewDevopsClient(&devops.DevopsConfig{
				PeerReportAPI: cctx.String("report-host"),
			})

			node := basenode.NewBasenode(config, client)
			if node == nil {
				return xerrors.Errorf("cannot init basenode")
			}

			rpcPeer := peer.NewPeer(node)
			if rpcPeer == nil {
				return xerrors.Errorf("cannot init peer")
			}

			rpcPeer.Run()

			ch := make(chan int)
			<-ch

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf(log.Fields{}, "fail to run %v: %v", app.Name, err)
	}
}
