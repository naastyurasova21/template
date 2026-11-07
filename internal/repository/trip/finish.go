package trip

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/db"
	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
)

func (r *Repo) Finish(ctx context.Context, id uuid.UUID) (*model.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.ExecutorFromContext(ctx, r.pool)
	now := time.Now().UTC()

	updateTrip, argsUpdate, err := sq.Update("trips").
		Set("status", string(trip.StatusCompleted)).
		Set("finished_at", now).
		Set("updated_at", now).
		Where(sq.Eq{"id": id, "status": string(trip.StatusActive)}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update trip: %w", err)
	}

	res, err := exec.Exec(ctx, updateTrip, argsUpdate...)
	if err != nil {
		return nil, fmt.Errorf("update trip: %w", err)
	}

	if res.RowsAffected() == 0 {
		if _, err := r.GetByID(ctx, id); err != nil {
			return nil, err
		}
		return nil, trip.ErrTripCompleted
	}

	insertHistory, argsHistory, err := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(id, string(trip.StatusActive), string(trip.StatusCompleted), "trip finished", now).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert history: %w", err)
	}

	if _, err := exec.Exec(ctx, insertHistory, argsHistory...); err != nil {
		return nil, fmt.Errorf("insert history: %w", err)
	}

	return r.GetByID(ctx, id)
}
