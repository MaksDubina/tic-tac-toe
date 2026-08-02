package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type JwtRequestExtension struct{}

func (e JwtRequestExtension) Sign(r *http.Request, uuidStr string) *http.Request {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return r
	}

	ctx := context.WithValue(r.Context(), userIDKey, parsedUUID)
	return r.WithContext(ctx)
}

func (e JwtRequestExtension) GetUUID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		return "", false
	}

	return id.String(), true
}
