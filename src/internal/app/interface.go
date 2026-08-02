package app

import (
	"context"
	"jwtAuth/src/internal/domain"

	"github.com/google/uuid"
)

type GameRepository interface {
	Save(ctx context.Context, game *domain.Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
	GetAvailableGames(ctx context.Context) ([]*domain.Game, error)
	GetFinishedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error)
	GetTopPlayersByWinRate(ctx context.Context, limit int) ([]*domain.LiederBoard, error)
}

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetById(ctx context.Context, id string) (*domain.User, error)
}

type TokenProvider interface {
	GenerateAccessToken(user domain.User) (string, error)
	GenerateRefreshToken(user domain.User) (string, error)
	ValidateAccessToken(tokenStr string) bool
	ValidateRefreshToken(tokenStr string) bool
	GetIdFromToken(tokenStr string) (string, error)
}
