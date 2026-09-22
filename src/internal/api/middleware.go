package api

import (
	"jwtAuth/src/internal/metrics"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UserAuthenticator struct {
	tokenProvider TokenProvider
	jwtExt        JwtRequestExtension
}

func NewUserAuthenticator(tp TokenProvider) *UserAuthenticator {
	return &UserAuthenticator{
		tokenProvider: tp,
		jwtExt:        JwtRequestExtension{}}
}

func (a *UserAuthenticator) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: invalid format", http.StatusUnauthorized)
			return
		}

		accessToken := parts[1]

		if !a.tokenProvider.ValidateAccessToken(accessToken) {
			http.Error(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)
			return
		}

		isValid := a.tokenProvider.ValidateAccessToken(accessToken)

		if !isValid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		uuid, err := a.tokenProvider.GetIdFromToken(accessToken)
		if err != nil {
			http.Error(w, "Unauthorized: identity error", http.StatusUnauthorized)
			return
		}

		r = a.jwtExt.Sign(r, uuid)

		next.ServeHTTP(w, r)
	})
}

func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()
			m.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
			m.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(ww.status)).Inc()
		})
	}
}

// statusRecorder нужен, чтобы перехватить статус-код ответа
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
