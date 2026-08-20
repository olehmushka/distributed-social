// Package search is the HTTP transport layer for the search service: a
// query handler plus the fx wiring that assembles the domain service in
// internal/search and registers its event consumers. There is no write
// handler here on purpose -- see internal/search's package doc.
package search

import (
	"context"
	"encoding/json"
	"fmt"

	logging "github.com/ipfs/go-log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	"github.com/olehmushka/distributed-social/internal/eventbus"
	"github.com/olehmushka/distributed-social/internal/eventsapi"
	searchdomain "github.com/olehmushka/distributed-social/internal/search"
	"github.com/olehmushka/distributed-social/server"
)

var Module = fx.Options(
	fx.Provide(
		newPool,
		newJetStream,
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
		return nil, fmt.Errorf("connect to search db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping search db: %w", err)
	}
	if err := searchdomain.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate search db: %w", err)
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

func newRepository(pool *pgxpool.Pool) searchdomain.Repository {
	return searchdomain.NewPostgresRepository(pool)
}

func newService(repo searchdomain.Repository, logger *logging.ZapEventLogger) *searchdomain.Service {
	return searchdomain.NewService(repo, logger)
}

type consumerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	JS        nats.JetStreamContext
	Service   *searchdomain.Service
	Logger    *logging.ZapEventLogger
}

// registerConsumers builds the search index entirely from events published
// by accounts and admins -- search never queries their databases directly,
// so it can be deployed, scaled, and rebuilt independently of them.
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

	if err := subscribe(eventsapi.SubjectPostCreated, "search-on-post-created", func(ctx context.Context, env eventbus.Envelope) error {
		var payload eventsapi.PostCreatedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return p.Service.HandlePostCreated(ctx, payload)
	}); err != nil {
		return err
	}

	if err := subscribe(eventsapi.SubjectPostRemoved, "search-on-post-removed", func(ctx context.Context, env eventbus.Envelope) error {
		var payload eventsapi.PostRemovedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return p.Service.HandlePostRemoved(ctx, payload)
	}); err != nil {
		return err
	}

	if err := subscribe(eventsapi.SubjectUserSuspended, "search-on-user-suspended", func(ctx context.Context, env eventbus.Envelope) error {
		var payload eventsapi.UserSuspendedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return err
		}
		return p.Service.HandleUserSuspended(ctx, payload)
	}); err != nil {
		return err
	}

	if err := subscribe(eventsapi.SubjectUserRestored, "search-on-user-restored", func(ctx context.Context, env eventbus.Envelope) error {
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
