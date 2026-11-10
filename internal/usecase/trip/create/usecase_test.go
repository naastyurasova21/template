package create_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
	"github.com/naastyurasova21/template/internal/usecase/trip/create"
)

func TestExecute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	mockTx := NewMockTxManager(ctrl)
	mockMetrics := NewMockMetrics(ctrl)

	uc := create.New(mockRepo, mockTx, mockMetrics, slog.New(slog.NewTextHandler(io.Discard, nil)))

	in := create.Input{
		UserID:   uuid.New(),
		DriverID: uuid.New(),
		StartLat: 55.75, StartLon: 37.61,
		EndLat: 59.93, EndLon: 30.31,
		Price: 145000,
	}

	mockTx.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockRepo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, m model.Trip) (*model.Trip, error) {
			if m.UserID != in.UserID {
				t.Errorf("UserID = %v, want %v", m.UserID, in.UserID)
			}
			if m.DriverID != in.DriverID {
				t.Errorf("DriverID = %v, want %v", m.DriverID, in.DriverID)
			}
			if m.Price != in.Price {
				t.Errorf("Price = %d, want %d", m.Price, in.Price)
			}
			if m.Status != string(trip.StatusActive) {
				t.Errorf("Status = %q, want %q", m.Status, trip.StatusActive)
			}
			m.ID = uuid.New()
			return &m, nil
		})
	mockMetrics.EXPECT().IncCreated(gomock.Any())

	saved, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	if saved == nil {
		t.Fatal("Execute: saved is nil")
	}
	if saved.Status != string(trip.StatusActive) {
		t.Errorf("saved.Status = %q, want %q", saved.Status, trip.StatusActive)
	}
}

func TestExecute_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	mockTx := NewMockTxManager(ctrl)
	mockMetrics := NewMockMetrics(ctrl)
	uc := create.New(mockRepo, mockTx, mockMetrics, slog.New(slog.NewTextHandler(io.Discard, nil)))

	in := create.Input{
		UserID:   uuid.Nil,
		DriverID: uuid.New(),
		Price:    100,
	}

	_, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var verr *trip.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected *trip.ValidationError, got %T", err)
	}
	if verr.Field != "user_id" {
		t.Errorf("Field = %q, want %q", verr.Field, "user_id")
	}
}

func TestExecute_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	mockTx := NewMockTxManager(ctrl)
	mockMetrics := NewMockMetrics(ctrl)
	uc := create.New(mockRepo, mockTx, mockMetrics, slog.New(slog.NewTextHandler(io.Discard, nil)))

	in := create.Input{
		UserID:   uuid.New(),
		DriverID: uuid.New(),
		StartLat: 0, StartLon: 0,
		EndLat: 0, EndLon: 0,
		Price: 100,
	}

	mockTx.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockRepo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil, trip.ErrDriverBusy)

	_, err := uc.Execute(context.Background(), in)
	if !errors.Is(err, trip.ErrDriverBusy) {
		t.Fatalf("err = %v, want ErrDriverBusy", err)
	}
}
