// Package runner connects the Kafka reading stream, the engine and the
// executor, and keeps the engine's rules in sync with the database.
package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"IDIG4110/rules-engine/internal/engine"
	"IDIG4110/rules-engine/internal/executor"
	"IDIG4110/shared/dto"
)

type Consumer interface {
	Run(ctx context.Context, handle func(ctx context.Context, key, value []byte) error) error
}

type Runner struct {
	consumer Consumer
	engine   *engine.Engine
	executor *executor.Executor
	reload   func(context.Context) error
	tick     time.Duration
	refresh  time.Duration
}

func New(c Consumer, e *engine.Engine, x *executor.Executor, reload func(context.Context) error, tick, refresh time.Duration) *Runner {
	return &Runner{consumer: c, engine: e, executor: x, reload: reload, tick: tick, refresh: refresh}
}

// Run consumes readings until ctx is cancelled. Alongside, it fires rules
// whose for_seconds has passed and reloads rules changed outside the API.
func (r *Runner) Run(ctx context.Context) error {
	go r.background(ctx)
	return r.consumer.Run(ctx, r.handle)
}

func (r *Runner) handle(ctx context.Context, _, value []byte) error {
	var reading dto.Reading
	if err := json.Unmarshal(value, &reading); err != nil {
		return fmt.Errorf("decoding reading: %w", err)
	}
	if reading.GatewayID == "" || reading.ExternalEntityID == "" || reading.Timestamp.IsZero() {
		return errors.New("reading is missing gateway_id, external_entity_id or timestamp")
	}
	for _, f := range r.engine.Handle(reading, time.Now()) {
		r.execute(ctx, f)
	}
	return nil
}

func (r *Runner) background(ctx context.Context) {
	tick := time.NewTicker(r.tick)
	refresh := time.NewTicker(r.refresh)
	defer tick.Stop()
	defer refresh.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-tick.C:
			for _, f := range r.engine.Tick(now) {
				r.execute(ctx, f)
			}
		case <-refresh.C:
			if err := r.reload(ctx); err != nil {
				slog.Error("Refreshing rules", "error", err)
			}
		}
	}
}

func (r *Runner) execute(ctx context.Context, f engine.Firing) {
	slog.Info("Rule fired", "rule", f.Rule.Name, "rule_id", f.Rule.ID, "entity", f.Reading.ExternalEntityID)
	if err := r.executor.Execute(ctx, f); err != nil {
		slog.Error("Executing rule actions", "rule", f.Rule.Name, "error", err)
	}
}
