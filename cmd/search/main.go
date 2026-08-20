package main

import (
	"os"

	"go.uber.org/fx"

	"github.com/carlmjohnson/versioninfo"
	logging "github.com/ipfs/go-log"
	"github.com/olehmushka/distributed-social/api/search"
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
		Name:    "search",
		Usage:   "builds a full-text index from the event stream and serves queries against it",
		Version: versioninfo.Short(),
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "api-listen",
			Value:   "0.0.0.0:9012",
			EnvVars: []string{"SEARCH_HOST"},
		},
		&cli.StringFlag{
			Name:    "db-dsn",
			Value:   "postgres://postgres:postgres@localhost:5435/search?sslmode=disable",
			EnvVars: []string{"SEARCH_DB_DSN"},
		},
		&cli.StringFlag{
			Name:    "nats-url",
			Value:   "nats://localhost:4222",
			EnvVars: []string{"NATS_URL"},
		},
	}

	app.Action = Run

	return app.Run(args)
}

func Run(cctx *cli.Context) error {
	fx.New(
		fx.Supply(server.Addr(cctx.String("api-listen"))),
		fx.Supply(server.Name("search")),
		fx.Supply(server.DBDSN(cctx.String("db-dsn"))),
		fx.Supply(server.NatsURL(cctx.String("nats-url"))),
		fx.Provide(func() *logging.ZapEventLogger { return logging.Logger("search") }),
		search.Module,
		server.Module,
	).Run()

	return nil
}
