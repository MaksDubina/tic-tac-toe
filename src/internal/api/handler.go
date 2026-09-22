package api

import (
	"encoding/json"
	"jwtAuth/src/internal/domain"
	"jwtAuth/src/internal/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type GameHandler struct {
	service     GameService
	authService AuthService
	jwtExt      JwtRequestExtension
	metrics     *metrics.Metrics
}

const AIPlayerUUID = "00000000-0000-0000-0000-000000000000"

func NewGameHandler(s GameService, a AuthService, m *metrics.Metrics) *GameHandler {
	return &GameHandler{
		service:     s,
		authService: a,
		jwtExt:      JwtRequestExtension{},
		metrics:     m,
	}
}

// CreateGame godoc
// @Summary      Создание новой игровой комнаты
// @Description  Инициализирует новую игру 3х3 с выбором оппонента (ИИ или человек). Требует авторизацию.
// @Tags         game
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body     CreateGameRequest  true  "Принимает строку ai или human"
// @Success      201     {object}  GameResponse
// @Failure      400     {object}  ErrorResponse      "Неверный формат тела запроса или невалидный opponent_type"
// @Failure      401     {object}  ErrorResponse      "Пользователь не авторизован"
// @Failure      500     {object}  ErrorResponse      "Внутренняя ошибка сервера при сохранении в БД"
// @Router       /game [post]
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		h.sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newGame := &domain.Game{
		ID:        uuid.New(),
		PlayerXID: &userID,
		Board: domain.Board{
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
		CreatedAt: time.Now(),
	}

	switch req.OpponentType {
	case "ai":
		aiUUID := uuid.MustParse(AIPlayerUUID)
		newGame.PlayerOID = &aiUUID
		newGame.Status = domain.PlayerXTurn

	case "human":
		newGame.PlayerOID = nil
		newGame.Status = domain.WaitingForPlayers

	default:
		h.sendError(w, "invalid opponent_type: must be 'ai' or 'human'", http.StatusBadRequest)
		return
	}

	if err := h.service.SaveGame(r.Context(), newGame); err != nil {
		h.sendError(w, "failed to save new game room", http.StatusInternalServerError)
		return
	}

	h.metrics.GamesCreatedTotal.Inc()
	h.metrics.GamesActive.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GameToResponse(newGame))
}

// HandleMove godoc
// @Summary      Совершение хода в игре
// @Description  Принимает новое состояние игрового поля от авторизованного пользователя.
// @Description  Проверяет очередность хода, валидирует корректность изменения клеток.
// @Description  Если оппонент — ИИ, то сервер автоматически рассчитывает и совершает ответный ход.
// @Tags         game
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        current_game_UUID  path      string       true  "UUID игровой сессии"
// @Param        request            body      GameRequest  true  "Новое состояние игрового поля"
// @Success      200                {object}  GameResponse "Ход успешно принят (и выполнен ответный ход ИИ, если применимо)"
// @Failure      400                {object}  ErrorResponse "Неверный формат UUID, сломанный JSON, игра окончена или еще не началась"
// @Failure      401                {object}  ErrorResponse "Пользователь не авторизован"
// @Failure      403                {object}  ErrorResponse "Ход не в свою очередь или пользователь не является участником этой игры"
// @Failure      404                {object}  ErrorResponse "Игровая сессия не найдена в базе данных"
// @Failure      422                {object}  ErrorResponse "Ошибка валидации доски (изменено более 1 клетки, затерты чужие фигуры)"
// @Failure      500                {object}  ErrorResponse "Внутренняя ошибка сервера при расчетах ИИ или сохранении в БД"
// @Router       /game/{current_game_UUID} [post]
func (h *GameHandler) HandleMove(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("current_game_UUID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(w, "invalid uuid format", http.StatusBadRequest)
		return
	}

	var req GameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "failed to decode json", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		h.sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	prevGame, err := h.service.GetGameByID(r.Context(), id)
	if err != nil {
		h.sendError(w, "game session not found", http.StatusNotFound)
		return
	}

	switch prevGame.Status {
	case domain.PlayerXTurn:
		if prevGame.PlayerXID == nil || *prevGame.PlayerXID != userID {
			h.sendError(w, "it is not your turn (waiting for player X)", http.StatusForbidden)
			return
		}
	case domain.PlayerOTurn:
		if prevGame.PlayerOID == nil || *prevGame.PlayerOID != userID {
			h.sendError(w, "it is not your turn (waiting for player O)", http.StatusForbidden)
			return
		}
	case domain.WaitingForPlayers:
		h.sendError(w, "game has not started yet, waiting for second player", http.StatusBadRequest)
		return
	case domain.XWon, domain.OWon, domain.Draw:
		h.sendError(w, "game is already finished", http.StatusBadRequest)
		return
	default:
		h.sendError(w, "unknown game status", http.StatusInternalServerError)
		return
	}

	currentGame := ToDomain(req, prevGame)
	if err := h.service.ValidateBoard(r.Context(), currentGame, prevGame); err != nil {
		h.sendError(w, "validation failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if prevGame.Status == domain.PlayerXTurn {
		currentGame.Status = domain.PlayerOTurn
	} else {
		currentGame.Status = domain.PlayerXTurn
	}

	status := h.service.IsGameOver(currentGame.Board, currentGame.Status)
	var updatedGame *domain.Game = currentGame

	if status == domain.XWon || status == domain.OWon || status == domain.Draw {
		updatedGame.Status = status
	} else {
		aiUUID := uuid.MustParse(AIPlayerUUID)

		if currentGame.PlayerOID != nil && *currentGame.PlayerOID == aiUUID {
			updatedGame, err = h.service.ComputeNextMove(r.Context(), currentGame)
			if err != nil {
				h.sendError(w, "ai computation error", http.StatusInternalServerError)
				return
			}
			aiStatus := h.service.IsGameOver(updatedGame.Board, updatedGame.Status)
			if aiStatus == domain.XWon || aiStatus == domain.OWon || aiStatus == domain.Draw {
				updatedGame.Status = aiStatus
			} else {
				updatedGame.Status = domain.PlayerXTurn
			}
		} else {
			if prevGame.Status == domain.PlayerXTurn {
				updatedGame.Status = domain.PlayerOTurn
			} else {
				updatedGame.Status = domain.PlayerXTurn
			}
		}
	}

	if err := h.service.SaveGame(r.Context(), updatedGame); err != nil {
		h.sendError(w, "failed to save game state", http.StatusInternalServerError)
		return
	}

	h.metrics.MovesTotal.Inc()

	switch updatedGame.Status {
	case domain.XWon, domain.OWon, domain.Draw:
		h.metrics.GamesActive.Dec()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GameToResponse(updatedGame))
}

// GetGame godoc
// @Summary      Получение информации о конкретной игре
// @Description  Возвращает текущее состояние игровой сессии (доску, статус, ID игроков) по её UUID.
// @Tags         game
// @Produce      json
// @Security     BearerAuth
// @Param        current_game_UUID  path      string  true  "UUID игровой сессии"
// @Success      200                {object}  GameResponse
// @Failure      400                {object}  ErrorResponse  "Неверный формат UUID"
// @Failure      401                {object}  ErrorResponse  "Пользователь не авторизован"
// @Failure      404                {object}  ErrorResponse  "Игровая сессия не найдена"
// @Router       /game/{current_game_UUID} [get]
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("current_game_UUID")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(w, "invalid uuid format", http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGameByID(r.Context(), id)
	if err != nil {
		h.sendError(w, "game session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GameToResponse(game))
}

// GetAvailableGames godoc
// @Summary      Получение списка доступных игр
// @Description  Возвращает список всех созданных игровых комнат, находящихся в статусе ожидания второго игрока (WaitingForPlayers).
// @Tags         game
// @Produce      json
// @Security     BearerAuth
// @Success      200              {array}   GameResponse  "Список доступных комнат (может быть пустым [])"
// @Failure      401              {object}  ErrorResponse "Пользователь не авторизован"
// @Failure      500              {object}  ErrorResponse "Внутренняя ошибка сервера при запросе к БД"
// @Router       /games/available [get]
func (h *GameHandler) GetAvailableGames(w http.ResponseWriter, r *http.Request) {

	games, err := h.service.GetAvailableGames(r.Context())
	if err != nil {
		h.sendError(w, "failed to fetch available games", http.StatusInternalServerError)
		return
	}

	res := make([]GameResponse, 0, len(games))
	for _, game := range games {
		res = append(res, GameToResponse(game))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// GetUser godoc
// @Summary      Получение информации о пользователе
// @Description  Возвращает публичный профиль пользователя (ID и логин) по его UUID.
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Param        user_uuid  path      string  true  "UUID пользователя"
// @Success      200        {object}  UserResponse
// @Failure      400        {object}  ErrorResponse  "Неверный формат UUID"
// @Failure      401        {object}  ErrorResponse  "Пользователь не авторизован"
// @Failure      404        {object}  ErrorResponse  "Пользователь не найден"
// @Router       /users/{user_uuid} [get]
func (h *GameHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_uuid")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.sendError(w, "invalid uuid format", http.StatusBadRequest)
		return
	}

	user, err := h.authService.GetById(r.Context(), id.String())
	if err != nil || user == nil {
		h.sendError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserToResponse(user))
}

// JoinGame godoc
// @Summary      Присоединение к существующей игре
// @Description  Позволяет авторизованному пользователю занять свободное место (Нолика) в комнате со статусом ожидания.
// @Tags         game
// @Produce      json
// @Security     BearerAuth
// @Param        current_game_UUID  path      string  true  "UUID игровой сессии"
// @Success      200                {object}  GameResponse   "Успешное подключение, статус игры изменен на PlayerXTurn"
// @Failure      400                {object}  ErrorResponse  "Неверный формат UUID или бизнес-ошибка (комната заполнена, попытка войти в свою комнату)"
// @Failure      401                {object}  ErrorResponse  "Пользователь не авторизован"
// @Router       /game/{current_game_UUID}/join [post]
func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
	if !ok {
		h.sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	gameID, err := uuid.Parse(r.PathValue("current_game_UUID"))
	if err != nil {
		h.sendError(w, "invalid uuid format", http.StatusBadRequest)
		return
	}

	updatedGame, err := h.service.JoinGameRoom(r.Context(), gameID, userID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GameToResponse(updatedGame))
}

func (h *GameHandler) sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

// GetGamesHistory godoc
// @Summary      Получение истории завершенных игр пользователя
// @Description  Возвращает список всех завершенных игр (победа, поражение, ничья), в которых участвовал текущий пользователь. Требует авторизацию.
// @Tags         game
// @Produce      json
// @Security     BearerAuth
// @Success      200     {array}   GameResponse  "Список завершенных игр"
// @Failure      401     {object}  ErrorResponse "Пользователь не авторизован"
// @Failure      500     {object}  ErrorResponse "Внутренняя ошибка сервера при чтении из БД"
// @Router       /games/history [get]
func (h *GameHandler) GetGamesHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuidStr, ok := h.jwtExt.GetUUID(r)
	if !ok {
		h.sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(uuidStr)
	if err != nil {
		h.sendError(w, "invalid user identity format", http.StatusBadRequest)
		return
	}

	finishedGames, err := h.service.GetGamesHistory(r.Context(), userID)
	if err != nil {
		h.sendError(w, "failed to fetch games history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var response []GameResponse
	for _, game := range finishedGames {
		response = append(response, GameToResponse(game))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetLiederBoard godoc
// @Summary      Получение таблицы лидеров
// @Description  Возвращает топ N игроков с наилучшим соотношением побед к поражениям и ничьим.
// @Tags         game
// @Produce      json
// @Security     BearerAuth
// @Param        limit query    int    false  "Количество игроков в выборке (по умолчанию 10)"
// @Success      200   {array}  domain.LiederBoard
// @Failure      400   {object}  ErrorResponse "Некорректное значение параметра limit"
// @Failure      500   {object}  ErrorResponse "Внутренняя ошибка сервера"
// @Router       /games/leaderboard [get]
func (h *GameHandler) GetLiederBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			h.sendError(w, "invalid limit parameter: must be a positive integer", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	leaderboard, err := h.service.GetLeaderboard(r.Context(), limit)
	if err != nil {
		h.sendError(w, "failed to fetch leaderboard: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(leaderboard)
}
