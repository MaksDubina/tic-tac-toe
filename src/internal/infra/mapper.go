package infra

import (
	"encoding/json"
	"jwtAuth/src/internal/domain"

	"github.com/google/uuid"
)

func ToModel(g *domain.Game) *GameModel {
	boardBytes, _ := json.Marshal(g.Board)
	return &GameModel{
		ID:        g.ID,
		BoardJSON: boardBytes,
		PlayerOID: g.PlayerOID,
		PlayerXID: g.PlayerXID,
		Status:    int(g.Status),
		CreatedAt: g.CreatedAt,
	}
}

func ToDomain(m *GameModel) *domain.Game {
	var board [][]int
	_ = json.Unmarshal(m.BoardJSON, &board)
	return &domain.Game{
		ID:        m.ID,
		Board:     board,
		PlayerOID: m.PlayerOID,
		PlayerXID: m.PlayerXID,
		Status:    domain.GameStatus(m.Status),
		CreatedAt: m.CreatedAt,
	}
}

func UserToModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:       u.ID.String(),
		Login:    u.Login,
		Password: u.Password,
	}
}

func UserToDomain(m *UserModel) *domain.User {
	id, _ := uuid.Parse(m.ID)
	return &domain.User{
		ID:       id,
		Login:    m.Login,
		Password: m.Password,
	}
}

func LeaderboardToDomain(l *LiederBoardModel) *domain.LiederBoard {
	id, _ := uuid.Parse(l.ID)
	return &domain.LiederBoard{
		ID:      id,
		WinRate: l.WinRate,
	}
}
