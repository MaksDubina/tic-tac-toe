package infra

import (
	"context"
	"errors"
	"fmt"
	"jwtAuth/src/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, g *domain.Game) error {
	model := ToModel(g)

	query := `
		INSERT INTO games (id, board, player_x_id, player_o_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) 
		DO UPDATE SET 
    		board = EXCLUDED.board,
    		status = EXCLUDED.status,
    		player_o_id = EXCLUDED.player_o_id;
	`

	_, err := r.pool.Exec(ctx, query, model.ID, model.BoardJSON, model.PlayerXID, model.PlayerOID, model.Status, model.CreatedAt)

	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	query := `SELECT id, board, player_x_id, player_o_id, status, created_at FROM games WHERE id = $1;`

	var model GameModel

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&model.ID,
		&model.BoardJSON,
		&model.PlayerXID,
		&model.PlayerOID,
		&model.Status,
		&model.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("game not found")
		}
		return nil, err
	}

	return ToDomain(&model), nil
}

func (r *Repository) GetAvailableGames(ctx context.Context) ([]*domain.Game, error) {

	query := `SELECT id, player_x_id, player_o_id, board, status, created_at FROM games WHERE status = $1;`

	rows, err := r.pool.Query(ctx, query, int(domain.WaitingForPlayers))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*domain.Game
	for rows.Next() {
		var model GameModel
		err := rows.Scan(&model.ID, &model.PlayerXID, &model.PlayerOID, &model.BoardJSON, &model.Status, &model.CreatedAt)
		if err != nil {
			return nil, err
		}
		games = append(games, ToDomain(&model))
	}

	return games, nil
}

func (r *Repository) GetFinishedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error) {
	query := `
		SELECT id, board, player_x_id, player_o_id, status, created_at 
		FROM games 
		WHERE (player_x_id = $1 OR player_o_id = $1) 
		  AND status IN ($2, $3, $4)
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query,
		userID,
		domain.XWon,
		domain.OWon,
		domain.Draw,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query finished games: %w", err)
	}
	defer rows.Close()

	var games []*domain.Game

	for rows.Next() {
		var model GameModel

		err := rows.Scan(
			&model.ID,
			&model.BoardJSON,
			&model.PlayerXID,
			&model.PlayerOID,
			&model.Status,
			&model.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game row: %w", err)
		}

		games = append(games, ToDomain(&model))

	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return games, nil
}

func (r *Repository) GetTopPlayersByWinRate(ctx context.Context, limit int) ([]*domain.LiederBoard, error) {
	query := `
	WITH user_stats AS (
		SELECT 
			user_id,
			COUNT(CASE WHEN (role = 'X' AND status = 4) OR (role = 'O' AND status = 5) THEN 1 END) as wins,
			COUNT(CASE WHEN (role = 'X' AND status = 5) OR (role = 'O' AND status = 4) THEN 1 END) as losses,
			COUNT(CASE WHEN status = 3 THEN 1 END) as draws
		FROM (
			SELECT player_x_id AS user_id, status, 'X'::text as role FROM games WHERE player_x_id IS NOT NULL AND status IN (3, 4, 5)
			UNION ALL
			SELECT player_o_id AS user_id, status, 'O'::text as role FROM games WHERE player_o_id IS NOT NULL AND status IN (3, 4, 5)
		) as matches
		GROUP BY user_id
	)
	SELECT 
		user_id,
		CASE 
			WHEN (losses + draws) = 0 THEN wins::float
			ELSE wins::float / (losses + draws)
		END as win_ratio
	FROM user_stats
	ORDER BY win_ratio DESC
	LIMIT $1;
`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute leaderboard query: %w", err)
	}
	defer rows.Close()

	var leaderboard []*domain.LiederBoard

	for rows.Next() {
		var item LiederBoardModel
		if err := rows.Scan(&item.ID, &item.WinRate); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %w", err)
		}
		leaderboard = append(leaderboard, LeaderboardToDomain(&item))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("leaderboard rows iteration error: %w", err)
	}

	return leaderboard, nil
}
