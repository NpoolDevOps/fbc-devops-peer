package main

import (
	log "github.com/EntropyPool/entropy-logger"
	basenode "github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	fullminer "github.com/NpoolDevOps/fbc-devops-peer/fullminer"
	fullnode "github.com/NpoolDevOps/fbc-devops-peer/fullnode"
	gateway "github.com/NpoolDevOps/fbc-devops-peer/gateway"
	miner "github.com/NpoolDevOps/fbc-devops-peer/miner"
	node "github.com/NpoolDevOps/fbc-devops-peer/node"
	"github.com/NpoolDevOps/fbc-devops-peer/peer"
	storage "github.com/NpoolDevOps/fbc-devops-peer/storage"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	worker "github.com/NpoolDevOps/fbc-devops-peer/worker"
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
				Name: "report-host",
			},
			&cli.StringFlag{
				Name: "username",
			},
			&cli.StringFlag{
				Name: "password",
			},
			&cli.StringFlag{
				Name: "network-type",
			},
			&cli.BoolFlag{
				Name: "test-mode",
			},
		},
		Action: func(cctx *cli.Context) error {
			if cctx.String("main-role") == "" {
				return xerrors.Errorf("main-role is must")
			}

			if cctx.String("network-type") == "" {
				return xerrors.Errorf("network type is must")
			}

			if cctx.String("username") == "" || cctx.String("password") == "" {
				return xerrors.Errorf("invalid username or password")
			}

			config := &basenode.BasenodeConfig{
				NodeConfig: &basenode.NodeConfig{
					MainRole: cctx.String("main-role"),
				},
				Username:    cctx.String("username"),
				Password:    cctx.String("password"),
				NetworkType: cctx.String("network-type"),
				TestMode:    cctx.Bool("test-mode"),
			}

			client := devops.NewDevopsClient(&devops.DevopsConfig{
				PeerReportAPI: cctx.String("report-host"),
				TestMode:      cctx.Bool("test-mode"),
			})

			var node node.Node

			switch cctx.String("main-role") {
			case types.GatewayNode:
				node = gateway.NewGatewayNode(config, client)
			case types.FullMinerNode:
				node = fullminer.NewFullMinerNode(config, client)
			case types.FullNode:
				node = fullnode.NewFullNode(config, client)
			case types.MinerNode:
				node = miner.NewMinerNode(config, client)
			case types.WorkerNode:
				node = worker.NewWorkerNode(config, client)
			case types.StorageNode:
				node = storage.NewStorageNode(config, client)
			}

			if node == nil {
				return xerrors.Errorf("cannot init basenode: %v", cctx.String("main-role"))
			}

			node.Banner()
			node.CreateExporter()

			rpcPeer := peer.NewPeer(node)
			if rpcPeer == nil {
				return xerrors.Errorf("cannot init peer")
			}

			node.SetPeer(rpcPeer)
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
