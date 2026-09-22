package app

import (
	"context"
	"errors"
	"fmt"
	"jwtAuth/src/internal/domain"

	"github.com/google/uuid"
)

type Service struct {
	repo GameRepository
}

const (
	Empty    = 0
	Player   = 1
	Computer = 2
)

func NewService(repo GameRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ValidateBoard(ctx context.Context, current, previous *domain.Game) error {
	if len(current.Board) != len(previous.Board) {
		return errors.New("размер поля изменился")
	}

	var expectedValue int
	switch previous.Status {
	case domain.PlayerXTurn:
		expectedValue = Player
	case domain.PlayerOTurn:
		expectedValue = Computer
	default:
		return errors.New("невозможный статус игры для совершения хода")
	}

	diffCount := 0
	for i := 0; i < len(current.Board); i++ {
		for j := 0; j < len(current.Board[i]); j++ {
			prevVal := previous.Board[i][j]
			currVal := current.Board[i][j]

			if prevVal != Empty && prevVal != currVal {
				return errors.New("изменение уже занятой клетки запрещено")
			}

			if prevVal == Empty && currVal != Empty {
				if currVal != expectedValue {
					return fmt.Errorf("неверный символ для текущего хода (ожидалось %d, получено %d)", expectedValue, currVal)
				}
				diffCount++
			}
		}
	}

	if diffCount != 1 {
		return errors.New("должен быть сделан ровно один ход")
	}

	return nil
}

func (s *Service) ComputeNextMove(ctx context.Context, game *domain.Game) (*domain.Game, error) {
	var compPiece, humanPiece int
	var compState, humanState domain.GameStatus

	if game.Status == domain.PlayerXTurn {
		compPiece = Player
		humanPiece = Computer
		compState = domain.XWon
		humanState = domain.OWon
	} else {
		compPiece = Computer
		humanPiece = Player
		compState = domain.OWon
		humanState = domain.XWon
	}

	bestScore := -1000
	var moveX, moveY int
	hasMove := false

	for i := 0; i < len(game.Board); i++ {
		for j := 0; j < len(game.Board[i]); j++ {
			if game.Board[i][j] == Empty {
				virtualBoard := copyBoard(game.Board)
				virtualBoard[i][j] = compPiece

				score := s.minimax(virtualBoard, 0, false, compState, humanState, compPiece, humanPiece)

				if score > bestScore {
					bestScore = score
					moveX, moveY = i, j
					hasMove = true
				}
			}
		}
	}

	if hasMove {
		game.Board[moveX][moveY] = compPiece
	}
	return game, nil
}

func (s *Service) minimax(board domain.Board, depth int, isMaximizing bool, compState, humanState domain.GameStatus, compPiece, humanPiece int) int {
	status := s.IsGameOver(board, compState)

	if status == domain.XWon || status == domain.OWon || status == domain.Draw {
		if status == compState {
			return 10 - depth
		}
		if status == humanState {
			return -10 + depth
		}
		return 0
	}

	if isMaximizing {
		bestScore := -1000
		for i := 0; i < len(board); i++ {
			for j := 0; j < len(board[i]); j++ {
				if board[i][j] == Empty {
					nextBoard := copyBoard(board)
					nextBoard[i][j] = compPiece

					score := s.minimax(nextBoard, depth+1, false, compState, humanState, compPiece, humanPiece)
					if score > bestScore {
						bestScore = score
					}
				}
			}
		}
		return bestScore
	} else {
		bestScore := 1000
		for i := 0; i < len(board); i++ {
			for j := 0; j < len(board[i]); j++ {
				if board[i][j] == Empty {
					nextBoard := copyBoard(board)
					nextBoard[i][j] = humanPiece

					score := s.minimax(nextBoard, depth+1, true, compState, humanState, compPiece, humanPiece)
					if score < bestScore {
						bestScore = score
					}
				}
			}
		}
		return bestScore
	}
}

func (s *Service) IsGameOver(board domain.Board, status domain.GameStatus) domain.GameStatus {
	size := len(board)

	for i := 0; i < size; i++ {
		if board[i][0] != Empty && allSame(board[i]) {
			if board[i][0] == 1 {
				return domain.XWon
			}
			return domain.OWon
		}
		col := make([]int, size)
		for j := 0; j < size; j++ {
			col[j] = board[j][i]
		}
		if col[0] != Empty && allSame(col) {
			if col[0] == 1 {
				return domain.XWon
			}
			return domain.OWon
		}
	}

	diag1 := make([]int, size)
	for i := 0; i < size; i++ {
		diag1[i] = board[i][i]
	}
	if diag1[0] != Empty && allSame(diag1) {
		if diag1[0] == 1 {
			return domain.XWon
		}
		return domain.OWon
	}

	diag2 := make([]int, size)
	for i := 0; i < size; i++ {
		diag2[i] = board[i][size-1-i]
	}
	if diag2[0] != Empty && allSame(diag2) {
		if diag2[0] == 1 {
			return domain.XWon
		}
		return domain.OWon
	}

	hasEmpty := false
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if board[i][j] == Empty {
				hasEmpty = true
				break
			}
		}
		if hasEmpty {
			break
		}
	}

	if hasEmpty {
		if status == domain.PlayerXTurn {
			return domain.PlayerOTurn
		}
		return domain.PlayerXTurn
	}

	return domain.Draw
}

func allSame(line []int) bool {
	for i := 1; i < len(line); i++ {
		if line[i] != line[0] {
			return false
		}
	}
	return true
}

func copyBoard(src domain.Board) domain.Board {
	dst := make(domain.Board, len(src))
	for i := range src {
		dst[i] = make([]int, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

func (s *Service) JoinGameRoom(ctx context.Context, gameID uuid.UUID, playerOID uuid.UUID) (*domain.Game, error) {
	game, err := s.repo.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if err := game.JoinPlayer(playerOID); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *Service) GetLeaderboard(ctx context.Context, limit int) ([]*domain.LiederBoard, error) {

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	leaderboard, err := s.repo.GetTopPlayersByWinRate(ctx, limit)
	if err != nil {
		return nil, err
	}

	return leaderboard, nil
}

func (s *Service) SaveGame(ctx context.Context, game *domain.Game) error {
	return s.repo.Save(ctx, game)
}

func (s *Service) GetGameByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAvailableGames(ctx context.Context) ([]*domain.Game, error) {
	return s.repo.GetAvailableGames(ctx)
}

func (s *Service) GetGamesHistory(ctx context.Context, userID uuid.UUID) ([]*domain.Game, error) {
	return s.repo.GetFinishedGamesByUserID(ctx, userID)
}
