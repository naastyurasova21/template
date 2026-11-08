package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/naastyurasova21/template/api"
)

func Recover(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					logger.Error("panic recovered",
						"panic", p,
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
					)

					detail := "Internal error"
					problem := api.Problem{
						Type:   "https://tripgo.example/problems/internal-error",
						Title:  "Internal Server Error",
						Status: http.StatusInternalServerError,
						Detail: &detail,
						Code:   "internal_error",
					}

					w.Header().Set("Content-Type", "application/problem+json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(problem)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
