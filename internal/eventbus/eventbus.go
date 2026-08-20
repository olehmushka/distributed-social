// Package eventbus wraps NATS JetStream so services can publish and
// consume durable events without depending on the NATS API directly in
// their domain code.
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	logging "github.com/ipfs/go-log"
	"github.com/nats-io/nats.go"
)

// StreamName is the single JetStream stream every service's events land on.
const StreamName = "SOCIAL_EVENTS"

// Envelope is the wire format for every event on the bus.
type Envelope struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	OccurredAt time.Time       `json:"occurredAt"`
	Payload    json.RawMessage `json:"payload"`
}

func NewEnvelope(eventType string, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal event payload: %w", err)
	}

	return Envelope{
		ID:         uuid.NewString(),
		Type:       eventType,
		OccurredAt: time.Now().UTC(),
		Payload:    raw,
	}, nil
}

// Connect opens a NATS connection and its JetStream context.
func Connect(url string) (*nats.Conn, nats.JetStreamContext, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
	if err != nil {
		return nil, nil, fmt.Errorf("connect to nats at %q: %w", url, err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("open jetstream context: %w", err)
	}

	return nc, js, nil
}

// EnsureStream creates the shared stream if it doesn't exist yet. Safe to
// call from every service on startup.
func EnsureStream(js nats.JetStreamContext, name string, subjects []string) error {
	if _, err := js.StreamInfo(name); err == nil {
		return nil
	}

	_, err := js.AddStream(&nats.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Retention: nats.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour,
		Storage:   nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("ensure jetstream stream %q: %w", name, err)
	}

	return nil
}

type Publisher interface {
	Publish(ctx context.Context, subject string, env Envelope) error
}

type natsPublisher struct {
	js nats.JetStreamContext
}

func NewPublisher(js nats.JetStreamContext) Publisher {
	return &natsPublisher{js: js}
}

func (p *natsPublisher) Publish(ctx context.Context, subject string, env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	if _, err := p.js.Publish(subject, data, nats.Context(ctx)); err != nil {
		return fmt.Errorf("publish to subject %q: %w", subject, err)
	}

	return nil
}

// Handler processes one event. Returning an error naks the message so
// JetStream redelivers it; handlers must therefore be idempotent.
type Handler func(ctx context.Context, env Envelope) error

// SubscribeDurable starts a background pull consumer bound to subject,
// dispatching every message to handler and acking on success. The returned
// stop func unsubscribes and stops the consume loop.
func SubscribeDurable(js nats.JetStreamContext, subject, durable string, handler Handler, logger *logging.ZapEventLogger) (func(), error) {
	sub, err := js.PullSubscribe(subject, durable, nats.ManualAck(), nats.AckWait(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("create durable consumer %q on %q: %w", durable, subject, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
			if err != nil {
				if err != nats.ErrTimeout && ctx.Err() == nil {
					logger.Warnw("jetstream fetch failed", "subject", subject, "durable", durable, "err", err)
				}
				continue
			}

			for _, msg := range msgs {
				var env Envelope
				if err := json.Unmarshal(msg.Data, &env); err != nil {
					logger.Errorw("dropping unparseable event", "subject", subject, "err", err)
					_ = msg.Ack() //nolint:errcheck
					continue
				}

				if err := handler(ctx, env); err != nil {
					logger.Warnw("event handler failed, will redeliver", "subject", subject, "durable", durable, "eventId", env.ID, "err", err)
					_ = msg.Nak() //nolint:errcheck
					continue
				}

				_ = msg.Ack() //nolint:errcheck
			}
		}
	}()

	return func() {
		cancel()
		_ = sub.Unsubscribe() //nolint:errcheck
	}, nil
}
