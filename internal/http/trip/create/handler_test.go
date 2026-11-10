package create_test

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/http/trip/create"
	"github.com/naastyurasova21/template/internal/model"
)

func TestCreateTrip_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := create.New(mockUC, slog.Default())

	body := `{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c","driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5","start_point":{"latitude":59.9398,"longitude":30.3146},"end_point":{"latitude":59.929,"longitude":30.3626},"price":1450}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(&model.Trip{
		ID:             uuid.New(),
		UserID:         uuid.MustParse("5cb72c04-7650-45c9-a79b-bcdba0631e0c"),
		DriverID:       uuid.MustParse("8860b315-ec86-42eb-a17c-7c163d721ff5"),
		StartLatitude:  59.9398,
		StartLongitude: 30.3146,
		EndLatitude:    59.929,
		EndLongitude:   30.3626,
		Price:          1450,
		Status:         string(trip.StatusActive),
	}, nil)

	h.CreateTrip(rr, req, api.CreateTripParams{})

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if rr.Header().Get("Location") == "" {
		t.Error("Location header is missing")
	}
}

func TestCreateTrip_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := create.New(mockUC, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBufferString(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTrip(rr, req, api.CreateTripParams{})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateTrip_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := NewMockUsecase(ctrl)
	h := create.New(mockUC, slog.Default())

	body := `{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c","driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5","start_point":{"latitude":59.9398,"longitude":30.3146},"end_point":{"latitude":59.929,"longitude":30.3626},"price":1450}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, errors.New("db connection lost"))

	h.CreateTrip(rr, req, api.CreateTripParams{})

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if bytes.Contains(rr.Body.Bytes(), []byte("db connection lost")) {
		t.Error("response body contains internal error details")
	}
}
