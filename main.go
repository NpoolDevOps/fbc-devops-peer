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
		},
		Action: func(cctx *cli.Context) error {
			mainRole := cctx.String("main-role")
			subRole := cctx.String("sub-role")
			parentSpec := cctx.String("parent-spec")

			switch mainRole {
			case "fullnode":
				return NewFullnodePeer(subRole, parentSpec).Run()
			case "miner":
				return NewMinerPeer(subRole, parentSpec).Run()
			case "worker":
				return NewWorkerPeer(subRole, parentSpec).Run()
			case "storage":
				return NewStoragePeer(subRole, parentSpec).Run()
			default:
				return xerrors.Errorf("Unknow main role %v", mainRole)
			}
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf(log.Fields{}, "fail to run %v: %v", app.Name, err)
	}
}
