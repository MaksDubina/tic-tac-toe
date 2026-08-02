package infra

import (
	"context"
	"errors"
	"jwtAuth/src/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Save(ctx context.Context, u *domain.User) error {
	model := UserToModel(u)
	query := `INSERT INTO users (id, login, password) VALUES ($1, $2, $3);`
	_, err := r.pool.Exec(ctx, query, model.ID, model.Login, model.Password)
	return err
}

func (r *UserRepository) GetById(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, login, password FROM users WHERE id = $1;`
	var model UserModel
	err := r.pool.QueryRow(ctx, query, id).Scan(&model.ID, &model.Login, &model.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return UserToDomain(&model), nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := `SELECT id, login, password FROM users WHERE login = $1;`
	var model UserModel
	err := r.pool.QueryRow(ctx, query, login).Scan(&model.ID, &model.Login, &model.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return UserToDomain(&model), nil
}
