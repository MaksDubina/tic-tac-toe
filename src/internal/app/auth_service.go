package app

import (
	"context"
	"errors"
	"jwtAuth/src/internal/domain"
	"strings"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo    UserRepository
	jwtProvider TokenProvider
}

func NewAuthService(userRepo UserRepository, jp TokenProvider) *AuthService {
	return &AuthService{userRepo: userRepo, jwtProvider: jp}
}

func (s *AuthService) Register(ctx context.Context, req domain.JwtRequest) (bool, error) {

	if len(strings.TrimSpace(req.Login)) < 3 || len(strings.TrimSpace(req.Password)) <= 5 {
		return false, errors.New("invalid login or password length")
	}

	_, err := s.userRepo.GetByLogin(ctx, req.Login)
	if err == nil {
		return false, errors.New("user already exists")
	}

	newUser := &domain.User{
		ID:       uuid.New(),
		Login:    req.Login,
		Password: req.Password,
	}

	if err := s.userRepo.Save(ctx, newUser); err != nil {
		return false, err
	}

	return true, nil

}

func (s *AuthService) Login(ctx context.Context, req domain.JwtRequest) (domain.JwtResponse, error) {
	user, err := s.userRepo.GetByLogin(ctx, req.Login)
	if err != nil {
		return domain.JwtResponse{}, errors.New("invalid credentials")
	}

	if user.Password != req.Password {
		return domain.JwtResponse{}, errors.New("invalid credentials")
	}

	return s.buildTokens(*user)
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, rawRefreshToken string) (domain.JwtResponse, error) {
	if !s.jwtProvider.ValidateRefreshToken(rawRefreshToken) {
		return domain.JwtResponse{}, errors.New("invalid refresh token")
	}

	uuid, err := s.jwtProvider.GetIdFromToken(rawRefreshToken)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	user, err := s.userRepo.GetById(ctx, uuid)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	accessToken, err := s.jwtProvider.GenerateAccessToken(*user)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	return domain.JwtResponse{
		Type:         "Bearer",
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, rawRefreshToken string) (domain.JwtResponse, error) {
	if !s.jwtProvider.ValidateRefreshToken(rawRefreshToken) {
		return domain.JwtResponse{}, errors.New("invalid refresh token")
	}

	uuid, err := s.jwtProvider.GetIdFromToken(rawRefreshToken)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	user, err := s.userRepo.GetById(ctx, uuid)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	return s.buildTokens(*user)
}

func (s *AuthService) buildTokens(user domain.User) (domain.JwtResponse, error) {
	access, err := s.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return domain.JwtResponse{}, err
	}
	refresh, err := s.jwtProvider.GenerateRefreshToken(user)
	if err != nil {
		return domain.JwtResponse{}, err
	}

	return domain.JwtResponse{
		Type:         "Bearer",
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *AuthService) Save(ctx context.Context, user *domain.User) error {
	return s.userRepo.Save(ctx, user)
}
func (s *AuthService) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	return s.userRepo.GetByLogin(ctx, login)
}
func (s *AuthService) GetById(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetById(ctx, id)
}
