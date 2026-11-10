package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	TripsCreated   metric.Int64Counter
	TripsCompleted metric.Int64Counter
}

func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("trip-service")

	created, err := meter.Int64Counter(
		"tripgo.trips.created",
		metric.WithDescription("Number of created trips"),
	)
	if err != nil {
		return nil, err
	}

	completed, err := meter.Int64Counter(
		"tripgo.trips.completed",
		metric.WithDescription("Number of completed trips"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		TripsCreated:   created,
		TripsCompleted: completed,
	}, nil
}

func (m *Metrics) IncCreated(ctx context.Context) {
	m.TripsCreated.Add(ctx, 1)
}

func (m *Metrics) IncCompleted(ctx context.Context) {
	m.TripsCompleted.Add(ctx, 1)
}
