package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/naastyurasova21/template/internal/db"
	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
)

const pgUniqueViolation = "23505"

func (r *Repo) Save(ctx context.Context, t model.Trip) (*model.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.ExecutorFromContext(ctx, r.pool)

	insertTrip, argsTrip, err := sq.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude",
			"end_latitude", "end_longitude",
			"price", "status",
			"started_at", "created_at", "updated_at",
		).
		Values(
			t.ID, t.UserID, t.DriverID,
			t.StartLatitude, t.StartLongitude,
			t.EndLatitude, t.EndLongitude,
			t.Price, t.Status,
			t.StartedAt, t.CreatedAt, t.UpdatedAt,
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert trip: %w", err)
	}

	if _, err := exec.Exec(ctx, insertTrip, argsTrip...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, trip.ErrDriverBusy
		}
		return nil, fmt.Errorf("insert trip: %w", err)
	}

	now := time.Now().UTC()
	insertHistory, argsHistory, err := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(t.ID, nil, t.Status, "trip created", now).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert history: %w", err)
	}

	if _, err := exec.Exec(ctx, insertHistory, argsHistory...); err != nil {
		return nil, fmt.Errorf("insert history: %w", err)
	}

	return r.GetByID(ctx, t.ID)
}
