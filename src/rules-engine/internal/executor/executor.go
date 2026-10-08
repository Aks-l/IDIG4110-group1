// Package executor carries out the actions of a rule firing: it publishes
// device commands, stores incidents with their notifications, and announces
// new incidents.
package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"IDIG4110/rules-engine/internal/engine"
	"IDIG4110/rules-engine/internal/rule"
	"IDIG4110/rules-engine/internal/store"
	"IDIG4110/shared/dto"
	"IDIG4110/shared/kafka"

	"github.com/google/uuid"
)

type IncidentStore interface {
	CreateIncident(ctx context.Context, inc store.Incident, channels []string) (store.Incident, error)
	RecordFiring(ctx context.Context, ruleID string, at time.Time) error
}

type Publisher interface {
	PublishJSON(ctx context.Context, topic, key string, v any) error
}

type Executor struct {
	store      IncidentStore
	publisher  Publisher
	commandTTL time.Duration
}

// New returns an executor. commandTTL is how long a command may wait before
// the gateway must refuse to run it (expires_at in gateway-api.md).
func New(s IncidentStore, p Publisher, commandTTL time.Duration) *Executor {
	return &Executor{store: s, publisher: p, commandTTL: commandTTL}
}

// Execute runs every action of the firing in order. A failing action does not
// stop the others; all errors are returned together.
func (x *Executor) Execute(ctx context.Context, f engine.Firing) error {
	var errs []error
	for _, a := range f.Rule.Actions {
		var err error
		switch a.Type {
		case rule.ActionCommand:
			err = x.sendCommand(ctx, f, a)
		case rule.ActionIncident:
			err = x.raiseIncident(ctx, f, a)
		case rule.ActionNotify:
			// Delivered together with the incident (see raiseIncident).
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("rule %q, %s action: %w", f.Rule.Name, a.Type, err))
		}
	}
	if err := x.store.RecordFiring(ctx, f.Rule.ID, f.At); err != nil {
		errs = append(errs, fmt.Errorf("recording firing of rule %q: %w", f.Rule.Name, err))
	}
	return errors.Join(errs...)
}

func (x *Executor) sendCommand(ctx context.Context, f engine.Firing, a rule.Action) error {
	gatewayID := a.GatewayID
	if gatewayID == "" {
		gatewayID = f.Reading.GatewayID
	}
	expires := f.At.Add(x.commandTTL)
	cmd := dto.Command{
		ID:               uuid.NewString(),
		GatewayID:        gatewayID,
		ExternalEntityID: a.Entity,
		Command:          a.Command,
		Parameters:       a.Parameters,
		IssuedAt:         f.At,
		ExpiresAt:        &expires,
		IssuedBy:         "rules-engine",
	}
	return x.publisher.PublishJSON(ctx, kafka.TopicCommands, dto.EntityKey(gatewayID, a.Entity), cmd)
}

func (x *Executor) raiseIncident(ctx context.Context, f engine.Firing, a rule.Action) error {
	value := readingValue(f.Reading)
	message := strings.NewReplacer("{{entity}}", f.Reading.ExternalEntityID, "{{value}}", value).Replace(a.Message)
	channels := f.Rule.Channels()

	saved, err := x.store.CreateIncident(ctx, store.Incident{
		RuleID:   f.Rule.ID,
		HomeID:   f.Rule.HomeID,
		Severity: a.Severity,
		Message:  message,
		Context: map[string]any{
			"rule":       f.Rule.Name,
			"gateway_id": f.Reading.GatewayID,
			"entity":     f.Reading.ExternalEntityID,
			"value":      value,
			"reading_at": f.Reading.Timestamp,
		},
		TriggeredAt: f.At,
	}, channels)
	if err != nil {
		return err
	}

	return x.publisher.PublishJSON(ctx, kafka.TopicIncidents, saved.HomeID, dto.IncidentEvent{
		ID:          saved.ID,
		RuleID:      saved.RuleID,
		HomeID:      saved.HomeID,
		Severity:    saved.Severity,
		Message:     saved.Message,
		Channels:    channels,
		Context:     saved.Context,
		TriggeredAt: saved.TriggeredAt,
	})
}

func readingValue(r dto.Reading) string {
	if r.ValueNum != nil {
		return strconv.FormatFloat(*r.ValueNum, 'f', -1, 64)
	}
	if r.ValueText != nil {
		return *r.ValueText
	}
	return ""
}
