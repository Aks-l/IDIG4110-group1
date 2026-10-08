package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"IDIG4110/rules-engine/internal/engine"
	"IDIG4110/rules-engine/internal/rule"
	"IDIG4110/rules-engine/internal/store"
	"IDIG4110/shared/dto"
	"IDIG4110/shared/kafka"
)

type fakeStore struct {
	incidents []store.Incident
	channels  []string
	fired     []string
	failWith  error
}

func (s *fakeStore) CreateIncident(_ context.Context, inc store.Incident, channels []string) (store.Incident, error) {
	if s.failWith != nil {
		return inc, s.failWith
	}
	inc.ID = "inc-1"
	s.incidents = append(s.incidents, inc)
	s.channels = channels
	return inc, nil
}

func (s *fakeStore) RecordFiring(_ context.Context, ruleID string, _ time.Time) error {
	s.fired = append(s.fired, ruleID)
	return nil
}

type published struct {
	topic, key string
	value      any
}

type fakePublisher struct{ records []published }

func (p *fakePublisher) PublishJSON(_ context.Context, topic, key string, v any) error {
	p.records = append(p.records, published{topic, key, v})
	return nil
}

var at = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func smokeFiring() engine.Firing {
	state := "on"
	return engine.Firing{
		At: at,
		Reading: dto.Reading{
			GatewayID: "11111111-1111-1111-1111-111111111111", ExternalEntityID: "binary_sensor.kitchen_smoke",
			Timestamp: at, ValueText: &state,
		},
		Rule: rule.Rule{
			ID: "rule-1", HomeID: "home-1", Name: "Smoke",
			Actions: []rule.Action{
				{Type: rule.ActionCommand, Entity: "lock.front_door", Command: "unlock"},
				{Type: rule.ActionIncident, Severity: "critical", Message: "Smoke at {{entity}} ({{value}})"},
				{Type: rule.ActionNotify, Channels: []string{"ui", "push"}},
			},
		},
	}
}

func TestExecuteRunsAllActions(t *testing.T) {
	s, p := &fakeStore{}, &fakePublisher{}
	if err := New(s, p, 30*time.Second).Execute(context.Background(), smokeFiring()); err != nil {
		t.Fatal(err)
	}

	if len(p.records) != 2 {
		t.Fatalf("published %d records, want command + incident event", len(p.records))
	}
	cmd, ok := p.records[0].value.(dto.Command)
	if !ok || p.records[0].topic != kafka.TopicCommands {
		t.Fatalf("first record = %+v, want a command on %s", p.records[0], kafka.TopicCommands)
	}
	if cmd.GatewayID != "11111111-1111-1111-1111-111111111111" || cmd.Command != "unlock" || cmd.ID == "" {
		t.Fatalf("command = %+v; gateway should default to the reading's", cmd)
	}
	if cmd.ExpiresAt == nil || !cmd.ExpiresAt.Equal(at.Add(30*time.Second)) {
		t.Fatalf("expires_at = %v, want issued_at + TTL", cmd.ExpiresAt)
	}

	if len(s.incidents) != 1 || s.incidents[0].Message != "Smoke at binary_sensor.kitchen_smoke (on)" {
		t.Fatalf("incidents = %+v", s.incidents)
	}
	if len(s.channels) != 2 {
		t.Fatalf("notification channels = %v, want ui and push", s.channels)
	}
	if p.records[1].topic != kafka.TopicIncidents {
		t.Fatalf("second record on %s, want %s", p.records[1].topic, kafka.TopicIncidents)
	}
	if len(s.fired) != 1 {
		t.Fatalf("firing not recorded")
	}
}

func TestExecuteContinuesAfterAFailedAction(t *testing.T) {
	s, p := &fakeStore{failWith: errors.New("db down")}, &fakePublisher{}
	err := New(s, p, time.Minute).Execute(context.Background(), smokeFiring())
	if err == nil {
		t.Fatal("expected the incident error to be returned")
	}
	if len(p.records) != 1 || len(s.fired) != 1 {
		t.Fatalf("command should still be sent and firing recorded: records=%d fired=%d", len(p.records), len(s.fired))
	}
}
