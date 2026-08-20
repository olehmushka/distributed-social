package admins

import (
	"context"
	"fmt"

	logging "github.com/ipfs/go-log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	admindomain "github.com/olehmushka/distributed-social/internal/admins"
	"github.com/olehmushka/distributed-social/internal/eventbus"
	"github.com/olehmushka/distributed-social/internal/eventsapi"
	"github.com/olehmushka/distributed-social/server"
)

var Module = fx.Options(
	fx.Provide(
		newPool,
		newJetStream,
		newPublisher,
		newRepository,
		newService,
		NewHandlers,
		NewRoutes,
	),
)

func newPool(lc fx.Lifecycle, dsn server.DBDSN) (*pgxpool.Pool, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, string(dsn))
	if err != nil {
		return nil, fmt.Errorf("connect to admins db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping admins db: %w", err)
	}
	if err := admindomain.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate admins db: %w", err)
	}

	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		pool.Close()
		return nil
	}})

	return pool, nil
}

func newJetStream(lc fx.Lifecycle, url server.NatsURL) (nats.JetStreamContext, error) {
	nc, js, err := eventbus.Connect(string(url))
	if err != nil {
		return nil, err
	}
	if err := eventbus.EnsureStream(js, eventbus.StreamName, eventsapi.StreamSubjects); err != nil {
		nc.Close()
		return nil, err
	}

	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		nc.Close()
		return nil
	}})

	return js, nil
}

func newPublisher(js nats.JetStreamContext) eventbus.Publisher {
	return eventbus.NewPublisher(js)
}

func newRepository(pool *pgxpool.Pool) admindomain.Repository {
	return admindomain.NewPostgresRepository(pool)
}

func newService(repo admindomain.Repository, pub eventbus.Publisher, logger *logging.ZapEventLogger) *admindomain.Service {
	return admindomain.NewService(repo, pub, logger)
}
