package api

import "time"

type GameRequest struct {
	Board [][]int `json:"board"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateGameRequest struct {
	OpponentType string `json:"opponent_type"`
}

type GameResponse struct {
	ID         string    `json:"id"`
	PlayerXID  *string   `json:"player_x_id"`
	PlayerOID  *string   `json:"player_o_id"`
	Board      [][]int   `json:"board"`
	GameStatus string    `json:"game_status"`
	CreatedAt  time.Time `json:"created_at"`
}
