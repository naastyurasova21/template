package get_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/http/trip/get"
	"github.com/naastyurasova21/template/internal/model"
)

func TestGetTrip_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := get.New(mockUC, slog.Default())

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	rr := httptest.NewRecorder()

	mockUC.EXPECT().Execute(gomock.Any(), id).Return(&model.Trip{
		ID:     id,
		Status: string(trip.StatusActive),
	}, nil)

	h.GetTrip(rr, req, id)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGetTrip_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := get.New(mockUC, slog.Default())

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	rr := httptest.NewRecorder()

	mockUC.EXPECT().Execute(gomock.Any(), id).Return(nil, trip.ErrTripNotFound)

	h.GetTrip(rr, req, id)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
