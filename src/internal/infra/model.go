package infra

import (
	"time"

	"github.com/google/uuid"
)

type GameModel struct {
	ID        uuid.UUID  `db:"id"`
	PlayerXID *uuid.UUID `db:"player_x_id"`
	PlayerOID *uuid.UUID `db:"player_o_id"`
	BoardJSON []byte     `db:"board"`
	Status    int        `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
}

type LiederBoardModel struct {
	ID      string  `db:"id"`
	WinRate float64 `db:"win_rate"`
}
