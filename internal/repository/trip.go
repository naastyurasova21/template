package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/db"
)

const pgUniqueViolation = "23505"

type TripRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

func (r *TripRepository) Create(ctx context.Context, data api.TripData) (*api.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.ExecutorFromContext(ctx, r.pool)

	id := uuid.New()
	now := time.Now().UTC()

	insertTrip, argsTrip, err := sq.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude",
			"end_latitude", "end_longitude",
			"price", "status",
			"started_at", "created_at", "updated_at",
		).
		Values(
			id, data.UserId, data.DriverId,
			data.StartPoint.Latitude, data.StartPoint.Longitude,
			data.EndPoint.Latitude, data.EndPoint.Longitude,
			data.Price, api.Active,
			now, now, now,
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert trip: %w", err)
	}

	if _, err := exec.Exec(ctx, insertTrip, argsTrip...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, ErrDriverBusy
		}
		return nil, fmt.Errorf("insert trip: %w", err)
	}

	insertHistory, argsHistory, err := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(id, nil, api.Active, "trip created", now).
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

func (r *TripRepository) GetByID(ctx context.Context, id uuid.UUID) (*api.Trip, error) {
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

	trip, err := scanTrip(exec.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("select trip: %w", err)
	}
	return trip, nil
}

func (r *TripRepository) Finish(ctx context.Context, id uuid.UUID) (*api.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	exec := db.ExecutorFromContext(ctx, r.pool)
	now := time.Now().UTC()

	// UPDATE trips
	updateTrip, argsUpdate, err := sq.Update("trips").
		Set("status", api.Completed).
		Set("finished_at", now).
		Set("updated_at", now).
		Where(sq.Eq{"id": id, "status": api.Active}).
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
		// Либо поездки нет, либо уже завершена
		if _, err := r.GetByID(ctx, id); err != nil {
			return nil, err // ErrTripNotFound
		}
		return nil, ErrTripAlreadyCompleted
	}

	// INSERT trip_status_history
	insertHistory, argsHistory, err := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(id, api.Active, api.Completed, "trip finished", now).
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
