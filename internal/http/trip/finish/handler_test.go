package finish_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/http/trip/finish"
	"github.com/naastyurasova21/template/internal/model"
)

func TestFinishTrip_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := finish.New(mockUC, slog.Default())

	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
	rr := httptest.NewRecorder()

	mockUC.EXPECT().Execute(gomock.Any(), id).Return(&model.Trip{
		ID:     id,
		Status: string(trip.StatusCompleted),
	}, nil)

	h.FinishTrip(rr, req, id)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestFinishTrip_AlreadyCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := finish.New(mockUC, slog.Default())

	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
	rr := httptest.NewRecorder()

	mockUC.EXPECT().Execute(gomock.Any(), id).Return(nil, trip.ErrTripCompleted)

	h.FinishTrip(rr, req, id)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusConflict)
	}
}
