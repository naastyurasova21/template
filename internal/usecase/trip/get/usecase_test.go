package get_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
	"github.com/naastyurasova21/template/internal/usecase/trip/get"
)

func TestExecute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	uc := get.New(mockRepo)

	id := uuid.New()
	expected := &model.Trip{ID: id, Status: string(trip.StatusActive)}

	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(expected, nil)

	got, err := uc.Execute(context.Background(), id)
	if err != nil {
		t.Fatalf("Execute: unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("got %p, want %p", got, expected)
	}
}

func TestExecute_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTripRepository(ctrl)
	uc := get.New(mockRepo)

	id := uuid.New()
	mockRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, trip.ErrTripNotFound)

	_, err := uc.Execute(context.Background(), id)
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("err = %v, want ErrTripNotFound", err)
	}
}
