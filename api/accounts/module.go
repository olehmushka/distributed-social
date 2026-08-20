// Package accounts is the HTTP transport layer for the accounts service:
// handlers, routes, and the fx wiring that assembles the domain service
// in internal/accounts with its Postgres pool and event-bus connection.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"

	logging "github.com/ipfs/go-log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	accountsdomain "github.com/olehmushka/distributed-social/internal/accounts"
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
	fx.Invoke(registerConsumers),
)

func newPool(lc fx.Lifecycle, dsn server.DBDSN) (*pgxpool.Pool, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, string(dsn))
	if err != nil {
		return nil, fmt.Errorf("connect to accounts db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping accounts db: %w", err)
	}
	if err := accountsdomain.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate accounts db: %w", err)
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

func newRepository(pool *pgxpool.Pool) accountsdomain.Repository {
	return accountsdomain.NewPostgresRepository(pool)
}

func newService(repo accountsdomain.Repository, pub eventbus.Publisher, logger *logging.ZapEventLogger) *accountsdomain.Service {
	return accountsdomain.NewService(repo, pub, logger)
}

type consumerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	JS        nats.JetStreamContext
	Service   *accountsdomain.Service
	Logger    *logging.ZapEventLogger
}

// registerConsumers subscribes accounts to the moderation decisions made by
// the admins service, so accounts stays the source of truth for status
// even though it never touches admins' database directly.
func registerConsumers(p consumerParams) error {
	var stops []func()

	subscribe := func(subject, durable string, handler eventbus.Handler) error {
		stop, err := eventbus.SubscribeDurable(p.JS, subject, durable, handler, p.Logger)
		if err != nil {
			return fmt.Errorf("subscribe %q: %w", subject, err)
		}
		stops = append(stops, stop)
		return nil
	}

	if err := subscribe(eventsapi.SubjectPostRemoved, "accounts-on-post-removed", func(ctx context.Context, env eventbus.Envelope) error {
		var payload eventsapi.PostRemovedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return p.Service.HandlePostRemoved(ctx, payload)
	}); err != nil {
		return err
	}

	if err := subscribe(eventsapi.SubjectUserSuspended, "accounts-on-user-suspended", func(ctx context.Context, env eventbus.Envelope) error {
		var payload eventsapi.UserSuspendedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return p.Service.HandleUserSuspended(ctx, payload)
	}); err != nil {
		return err
	}

	if err := subscribe(eventsapi.SubjectUserRestored, "accounts-on-user-restored", func(ctx context.Context, env eventbus.Envelope) error {
		var payload eventsapi.UserRestoredPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return p.Service.HandleUserRestored(ctx, payload)
	}); err != nil {
		return err
	}

	p.Lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		for _, stop := range stops {
			stop()
		}
		return nil
	}})

	return nil
}
