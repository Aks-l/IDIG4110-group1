package server

import (
	"context"

	"IDIG4110/shared/dto"
	"IDIG4110/shared/kafka"
)

// readingPublisher publishes normalized readings on twin.readings, keyed by
// gateway and entity so each entity's readings stay in order.
type readingPublisher struct {
	producer *kafka.Producer
}

func (p readingPublisher) PublishReading(ctx context.Context, r dto.Reading) error {
	return p.producer.PublishJSON(ctx, kafka.TopicReadings, r.Key(), r)
}
