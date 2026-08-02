package domain

import (
	"time"

	"github.com/google/uuid"
)

type Board [][]int

type Game struct {
	ID        uuid.UUID
	PlayerXID *uuid.UUID
	PlayerOID *uuid.UUID
	Board     Board
	Status    GameStatus

	CreatedAt time.Time
}

type GameStatus int

const (
	WaitingForPlayers GameStatus = iota
	PlayerXTurn
	PlayerOTurn
	Draw
	XWon
	OWon
)

func (s GameStatus) String() string {
	states := [...]string{"WaitingForPlayers", "PlayerXTurn", "PlayerOTurn", "Draw", "XWon", "OWON"}

	if s < WaitingForPlayers || s > OWon {
		return "Unknown"
	}
	return states[s]
}

type LiederBoard struct {
	ID      uuid.UUID `json:"id"`
	WinRate float64   `json:"win_rate"`
}
