package trip

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Repo struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	dbDuration   metric.Float64Histogram
}

var _ Repository = (*Repo)(nil)

func New(pool *pgxpool.Pool, queryTimeout time.Duration) *Repo {
	meter := otel.Meter("trip-service")
	durationHist, _ := meter.Float64Histogram(
		"db.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of database operations"),
	)

	return &Repo{
		pool:         pool,
		queryTimeout: queryTimeout,
		dbDuration:   durationHist,
	}
}

func (r *Repo) recordDBDuration(ctx context.Context, op string, start time.Time) {
	duration := time.Since(start).Seconds()
	r.dbDuration.Record(ctx, duration, metric.WithAttributes(
		attribute.String("db.operation.name", op),
		attribute.String("db.collection.name", "trips"),
	))
}
