package api

import (
	"net/http"
	"strings"
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
