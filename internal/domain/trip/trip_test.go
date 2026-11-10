package trip_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/domain/trip"
)

func validNew(t *testing.T) *trip.Trip {
	t.Helper()
	tr, err := trip.New(
		uuid.New(), uuid.New(),
		55.75, 37.61, 59.93, 30.31,
		145000,
		time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	return tr
}

func TestNew_Valid(t *testing.T) {
	tr := validNew(t)

	if tr.Status != trip.StatusActive {
		t.Errorf("Status = %q, want %q", tr.Status, trip.StatusActive)
	}
	if tr.FinishedAt != nil {
		t.Errorf("FinishedAt = %v, want nil", tr.FinishedAt)
	}
}

func TestNew_Validation(t *testing.T) {
	userID := uuid.New()
	driverID := uuid.New()
	now := time.Now()

	tests := []struct {
		name               string
		user               uuid.UUID
		driver             uuid.UUID
		startLat, startLon float64
		endLat, endLon     float64
		price              int64
		wantField          string
	}{
		{"empty user", uuid.Nil, driverID, 0, 0, 0, 0, 0, "user_id"},
		{"empty driver", userID, uuid.Nil, 0, 0, 0, 0, 0, "driver_id"},
		{"negative price", userID, driverID, 0, 0, 0, 0, -1, "price"},
		{"bad start lat", userID, driverID, 91, 0, 0, 0, 0, "start_point.latitude"},
		{"bad start lon", userID, driverID, 0, 181, 0, 0, 0, "start_point.longitude"},
		{"bad end lat", userID, driverID, 0, 0, -91, 0, 0, "end_point.latitude"},
		{"bad end lon", userID, driverID, 0, 0, 0, -181, 0, "end_point.longitude"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := trip.New(tt.user, tt.driver, tt.startLat, tt.startLon, tt.endLat, tt.endLon, tt.price, now)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var verr *trip.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("expected *trip.ValidationError, got %T", err)
			}
			if verr.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", verr.Field, tt.wantField)
			}
		})
	}
}

func TestFinish_Active(t *testing.T) {
	tr := validNew(t)
	finishedAt := time.Date(2026, 8, 24, 12, 30, 0, 0, time.UTC)

	if err := tr.Finish(finishedAt); err != nil {
		t.Fatalf("Finish: unexpected error: %v", err)
	}
	if tr.Status != trip.StatusCompleted {
		t.Errorf("Status = %q, want %q", tr.Status, trip.StatusCompleted)
	}
	if tr.FinishedAt == nil || !tr.FinishedAt.Equal(finishedAt) {
		t.Errorf("FinishedAt = %v, want %v", tr.FinishedAt, finishedAt)
	}
}

func TestFinish_AlreadyCompleted(t *testing.T) {
	tr := validNew(t)
	now := time.Now()
	if err := tr.Finish(now); err != nil {
		t.Fatalf("first Finish: %v", err)
	}

	firstFinishedAt := *tr.FinishedAt
	err := tr.Finish(now.Add(time.Hour))
	if !errors.Is(err, trip.ErrTripCompleted) {
		t.Fatalf("second Finish: err = %v, want ErrTripCompleted", err)
	}
	if !tr.FinishedAt.Equal(firstFinishedAt) {
		t.Errorf("FinishedAt was overwritten: %v, want %v", tr.FinishedAt, firstFinishedAt)
	}
}
