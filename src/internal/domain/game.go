package domain

import (
	"errors"

	"github.com/google/uuid"
)

func (g *Game) JoinPlayer(playerOID uuid.UUID) error {
	if g.Status != WaitingForPlayers {
		return errors.New("cannot join: room is not waiting for players")
	}
	if g.PlayerXID != nil && *g.PlayerXID == playerOID {
		return errors.New("cannot join your own room")
	}

	g.PlayerOID = &playerOID
	g.Status = PlayerXTurn
	return nil
}
