package commands

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"IDIG4110/shared/dto"
)

type fakeSink struct {
	topics []string
	qos    []byte
}

func (s *fakeSink) Publish(topic string, qos byte, _ []byte) error {
	s.topics = append(s.topics, topic)
	s.qos = append(s.qos, qos)
	return nil
}

var now = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func record(t *testing.T, expires time.Time) []byte {
	t.Helper()
	b, err := json.Marshal(dto.Command{
		ID: "c1", GatewayID: "11111111-1111-1111-1111-111111111111",
		ExternalEntityID: "valve.main_water", Command: "close",
		IssuedAt: now.Add(-time.Second), ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func forwarder(s Sink) *Forwarder {
	f := NewForwarder(s)
	f.now = func() time.Time { return now }
	return f
}

func TestForwardsValidCommandWithQoS1(t *testing.T) {
	sink := &fakeSink{}
	if err := forwarder(sink).Handle(context.Background(), nil, record(t, now.Add(time.Minute))); err != nil {
		t.Fatal(err)
	}
	if len(sink.topics) != 1 || sink.topics[0] != "twin/11111111-1111-1111-1111-111111111111/commands" || sink.qos[0] != 1 {
		t.Fatalf("published %v qos %v", sink.topics, sink.qos)
	}
}

func TestDropsExpiredCommand(t *testing.T) {
	sink := &fakeSink{}
	if err := forwarder(sink).Handle(context.Background(), nil, record(t, now.Add(-time.Second))); err != nil {
		t.Fatal(err)
	}
	if len(sink.topics) != 0 {
		t.Fatal("expired command must not be forwarded")
	}
}

func TestRejectsIncompleteCommand(t *testing.T) {
	if err := forwarder(&fakeSink{}).Handle(context.Background(), nil, []byte(`{"id":"c1"}`)); err == nil {
		t.Fatal("expected an error for a command without target")
	}
}
