package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/naastyurasova21/template/internal/db"
	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*model.Trip, error) {
	start := time.Now()
	defer r.recordDBDuration(ctx, "SELECT", start)
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.ExecutorFromContext(ctx, r.pool)

	sql, args, err := sq.Select(
		"id", "user_id", "driver_id",
		"start_latitude", "start_longitude",
		"end_latitude", "end_longitude",
		"price", "status",
		"started_at", "finished_at",
	).
		From("trips").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select trip: %w", err)
	}

	t, err := scanTrip(exec.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, trip.ErrTripNotFound
		}
		return nil, fmt.Errorf("select trip: %w", err)
	}
	return t, nil
}
