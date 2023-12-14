package main

import (
	"os"

	"go.uber.org/fx"

	"github.com/carlmjohnson/versioninfo"
	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/api/admins"
	"github.com/olehmushka/distributed-social/server"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

func main() {
	if err := run(os.Args); err != nil {
		slog.Error("exiting", "err", err)
		os.Exit(-1)
	}
}

func run(args []string) error {
	app := cli.App{
		Name:    "admins",
		Usage:   "serving admins information",
		Version: versioninfo.Short(),
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "api-listen",
			Value:   "0.0.0.0:9011",
			EnvVars: []string{"ADMINS_HOST"},
		},
	}

	app.Action = Run

	return app.Run(args)
}

func Run(cctx *cli.Context) error {
	fx.New(
		fx.Supply(server.Addr(cctx.String("api-listen"))),
		fx.Supply(server.Name("admins")),
		fx.Provide(func() *logging.ZapEventLogger { return logging.Logger("admins") }),
		admins.Module,
		server.Module,
	).Run()

	return nil
}
