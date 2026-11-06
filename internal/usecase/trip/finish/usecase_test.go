package finish_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
	"github.com/naastyurasova21/template/internal/usecase/trip/finish"
)

func TestExecute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	mockTx := NewMockTxManager(ctrl)
	uc := finish.New(mockRepo, mockTx)

	id := uuid.New()
	expected := &model.Trip{ID: id, Status: string(trip.StatusCompleted)}

	mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
	)
	mockRepo.EXPECT().Finish(gomock.Any(), id).Return(expected, nil)

	got, err := uc.Execute(context.Background(), id)
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("got %p, want %p", got, expected)
	}
}

func TestExecute_AlreadyCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	mockTx := NewMockTxManager(ctrl)
	uc := finish.New(mockRepo, mockTx)

	id := uuid.New()

	mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
	)
	mockRepo.EXPECT().Finish(gomock.Any(), id).Return(nil, trip.ErrTripCompleted)

	_, err := uc.Execute(context.Background(), id)
	if !errors.Is(err, trip.ErrTripCompleted) {
		t.Fatalf("err = %v, want ErrTripCompleted", err)
	}
}

func TestExecute_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	mockTx := NewMockTxManager(ctrl)
	uc := finish.New(mockRepo, mockTx)

	id := uuid.New()

	mockTx.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
	)
	mockRepo.EXPECT().Finish(gomock.Any(), id).Return(nil, trip.ErrTripNotFound)

	_, err := uc.Execute(context.Background(), id)
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("err = %v, want ErrTripNotFound", err)
	}
}
