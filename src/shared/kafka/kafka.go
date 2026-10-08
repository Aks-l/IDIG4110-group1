// Package kafka is a thin wrapper around franz-go for the platform's event
// bus (docs/decisions/0001-kafka-event-bus.md).
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	TopicReadings  = "twin.readings"
	TopicCommands  = "twin.commands"
	TopicIncidents = "twin.incidents"
)

// ParseBrokers splits a comma-separated broker list such as "kafka:29092".
func ParseBrokers(s string) []string {
	var brokers []string
	for _, b := range strings.Split(s, ",") {
		if b = strings.TrimSpace(b); b != "" {
			brokers = append(brokers, b)
		}
	}
	return brokers
}

type Producer struct {
	client *kgo.Client
}

func NewProducer(brokers []string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}
	return &Producer{client: client}, nil
}

// PublishJSON writes v as JSON to topic and waits for the broker to accept it.
// Records with the same key always land in the same partition, in order.
func (p *Producer) PublishJSON(ctx context.Context, topic, key string, v any) error {
	value, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding %s record: %w", topic, err)
	}
	record := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("publishing to %s: %w", topic, err)
	}
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}

type Consumer struct {
	client *kgo.Client
}

// NewConsumer joins group and consumes topics. A group that has never
// committed starts at the newest records, so a new service reacts to new
// events instead of replaying the whole history.
func NewConsumer(brokers []string, group string, topics ...string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}
	return &Consumer{client: client}, nil
}

// Run polls until ctx is cancelled and calls handle for every record. A
// handler error is logged and the record is skipped, so one bad message
// cannot stop the stream.
func (c *Consumer) Run(ctx context.Context, handle func(ctx context.Context, key, value []byte) error) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if fetches.IsClientClosed() {
			return nil
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			slog.Error("Kafka fetch", "topic", topic, "partition", partition, "error", err)
		})
		fetches.EachRecord(func(r *kgo.Record) {
			if err := handle(ctx, r.Key, r.Value); err != nil {
				slog.Error("Kafka record handler", "topic", r.Topic, "offset", r.Offset, "error", err)
			}
		})
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}
