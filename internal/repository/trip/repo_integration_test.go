//go:build integration

package trip_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/naastyurasova21/template/internal/config"
	"github.com/naastyurasova21/template/internal/db"
	domaintrip "github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
	repotrip "github.com/naastyurasova21/template/internal/repository/trip"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	err := godotenv.Load("../../../.env")
	if err != nil {
		t.Logf("warning: could not load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	pool, err := db.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect to db: %v", err)
	}

	_, err = pool.Exec(context.Background(), "TRUNCATE trips, trip_status_history CASCADE")
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	return pool
}

func TestRepo_SaveAndGet(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	repo := repotrip.New(pool, 3*time.Second)
	ctx := context.Background()

	now := time.Now().UTC()
	m := model.Trip{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		DriverID:       uuid.New(),
		StartLatitude:  55.75,
		StartLongitude: 37.61,
		EndLatitude:    59.93,
		EndLongitude:   30.31,
		Price:          1450,
		Status:         string(domaintrip.StatusActive),
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	saved, err := repo.Save(ctx, m)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID != m.ID {
		t.Errorf("saved.ID = %v, want %v", saved.ID, m.ID)
	}

	fetched, err := repo.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.Status != string(domaintrip.StatusActive) {
		t.Errorf("status = %q, want %q", fetched.Status, domaintrip.StatusActive)
	}

	finished, err := repo.Finish(ctx, m.ID)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if finished.Status != string(domaintrip.StatusCompleted) {
		t.Errorf("status = %q, want %q", finished.Status, domaintrip.StatusCompleted)
	}
}

func TestRepo_TransactionRollback(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	repo := repotrip.New(pool, 3*time.Second)
	ctx := context.Background()

	now := time.Now().UTC()
	driverID := uuid.New()

	m1 := model.Trip{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		DriverID:       driverID,
		StartLatitude:  55.75,
		StartLongitude: 37.61,
		EndLatitude:    59.93,
		EndLongitude:   30.31,
		Price:          100,
		Status:         string(domaintrip.StatusActive),
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if _, err := repo.Save(ctx, m1); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	m2 := m1
	m2.ID = uuid.New()

	_, err := repo.Save(ctx, m2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips WHERE driver_id = $1", driverID).Scan(&count)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 1 {
		t.Errorf("trips count = %d, want 1 (вторая поездка должна была откатиться)", count)
	}
}
