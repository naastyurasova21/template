package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/naastyurasova21/template/api"
)

var problemInfo = map[string]struct {
	Title string
	Slug  string
}{
	"invalid_request": {"Invalid request", "invalid-request"},
	"trip_not_found":  {"Trip not found", "trip-not-found"},
	"trip_completed":  {"Trip completed", "trip-completed"},
	"driver_busy":     {"Driver busy", "driver-busy"},
	"internal_error":  {"Internal Server Error", "internal-error"},
	"not_implemented": {"Not implemented", "not-implemented"},
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	info, ok := problemInfo[code]
	if !ok {
		info = struct {
			Title string
			Slug  string
		}{"Error", "error"}
	}

	instance := r.URL.Path

	problem := api.Problem{
		Type:     "https://tripgo.example/problems/" + info.Slug,
		Title:    info.Title,
		Status:   int32(status),
		Detail:   &detail,
		Instance: &instance,
		Code:     code,
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	var paramErr *api.InvalidParamFormatError
	if errors.As(err, &paramErr) {
		detail := "invalid parameter format"
		if paramErr.ParamName == "tripId" {
			detail = "invalid UUID format"
		}
		writeProblem(w, r, http.StatusBadRequest, "invalid_request", detail)
		return
	}
	writeProblem(w, r, http.StatusBadRequest, "invalid_request", "invalid request parameters")
}
