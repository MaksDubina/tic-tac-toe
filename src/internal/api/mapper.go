package api

import (
	"jwtAuth/src/internal/domain"
)

func ToDomain(req GameRequest, prevGame *domain.Game) *domain.Game {
	return &domain.Game{
		ID:        prevGame.ID,
		Board:     domain.Board(req.Board),
		PlayerXID: prevGame.PlayerXID,
		PlayerOID: prevGame.PlayerOID,
		CreatedAt: prevGame.CreatedAt,
	}
}

func GameToResponse(g *domain.Game) GameResponse {
	var xStr, oStr *string
	if g.PlayerXID != nil {
		s := g.PlayerXID.String()
		xStr = &s
	}
	if g.PlayerOID != nil {
		s := g.PlayerOID.String()
		oStr = &s
	}

	return GameResponse{
		ID:         g.ID.String(),
		PlayerXID:  xStr,
		PlayerOID:  oStr,
		Board:      [][]int(g.Board),
		GameStatus: g.Status.String(),
		CreatedAt:  g.CreatedAt,
	}
}

func UserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:    u.ID.String(),
		Login: u.Login,
	}
}
