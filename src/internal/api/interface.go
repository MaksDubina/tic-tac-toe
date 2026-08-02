package api

import (
	"context"
	"jwtAuth/src/internal/domain"

	"github.com/google/uuid"
)

type GameService interface {
	ComputeNextMove(ctx context.Context, game *domain.Game) (*domain.Game, error)
	ValidateBoard(ctx context.Context, current, previous *domain.Game) error
	IsGameOver(board domain.Board, status domain.GameStatus) domain.GameStatus
	JoinGameRoom(ctx context.Context, gameID uuid.UUID, playerOID uuid.UUID) (*domain.Game, error)
	GetLeaderboard(ctx context.Context, limit int) ([]*domain.LiederBoard, error)

	SaveGame(ctx context.Context, game *domain.Game) error
	GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
	GetAvailableGames(ctx context.Context) ([]*domain.Game, error)
	GetGamesHistory(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error)
}

type AuthService interface {
	Register(ctx context.Context, req domain.JwtRequest) (bool, error)
	Login(ctx context.Context, req domain.JwtRequest) (domain.JwtResponse, error)
	RefreshAccessToken(ctx context.Context, rawRefreshToken string) (domain.JwtResponse, error)
	RefreshTokens(ctx context.Context, rawRefreshToken string) (domain.JwtResponse, error)

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
