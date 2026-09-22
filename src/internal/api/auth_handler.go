package api

import (
	"encoding/json"
	"jwtAuth/src/internal/domain"
	"jwtAuth/src/internal/metrics"
	"net/http"
)

type AuthHandler struct {
	authService AuthService
	jwtExt      JwtRequestExtension
	metrics     *metrics.Metrics
}

func NewAuthHandler(as AuthService, m *metrics.Metrics) *AuthHandler {

	return &AuthHandler{
		authService: as,
		jwtExt:      JwtRequestExtension{},
		metrics:     m,
	}
}

// HandleRegister godoc
// @Summary      Регистрация нового пользователя
// @Description  Создает новый аккаунт и сразу возвращает JWT-токены для авторизации.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     SignUpRequest  true  "Данные для регистрации (логин и пароль)"
// @Success      201     {object}  AuthResponse   "Успешная регистрация"
// @Failure      400     {object}  AuthResponse   "Некорректное тело запроса (невалидный JSON)"
// @Failure      422     {object}  AuthResponse   "Ошибка бизнес-валидации (пользователь уже существует, слишком короткий пароль)"
// @Router       /auth/signup [post]
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendJSON(w, http.StatusBadRequest, AuthResponse{Error: "invalid body"})
		return
	}

	success, err := h.authService.Register(r.Context(), domain.JwtRequest(req))
	if err != nil {
		h.sendJSON(w, http.StatusUnprocessableEntity, AuthResponse{Error: err.Error()})
		return
	}

	h.metrics.AuthAttemptsTotal.WithLabelValues("signup", "success").Inc()

	h.sendJSON(w, http.StatusCreated, AuthResponse{Success: success})
}

// HandleLogin godoc
// @Summary      Авторизация пользователя (Вход)
// @Description  Проверяет учетные данные пользователя (логин/пароль) и возвращает пару JWT токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     domain.JwtRequest  true  "Данные для входа"
// @Success      200     {object}  domain.JwtResponse   "Успешный вход, возвращает токены"
// @Failure      400     {object}  domain.JwtResponse         "Некорректный формат JSON"
// @Failure      401     {object}  domain.JwtResponse         "Неверный логин или пароль"
// @Router       /auth/signin [post]
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.JwtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	JwtResponse, err := h.authService.Login(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.metrics.AuthAttemptsTotal.WithLabelValues("signin", "success").Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JwtResponse)
}

// RefreshAccess godoc
// @Summary      Обновление Access токена
// @Description  Принимает refreshToken и возвращает новый accessToken БЕЗ изменения сессии рефреша. Доступен без авторизации.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     domain.RefreshJwtRequest  true  "Действующий Refresh токен"
// @Success      200     {object}  domain.JwtResponse        "Новый access токен"
// @Failure      400     {object}  AuthResponse              "Некорректный формат JSON"
// @Failure      401     {object}  AuthResponse              "Невалидный или просроченный токен"
// @Router       /auth/refresh-access [post]
func (h *AuthHandler) RefreshAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.RefreshJwtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RefreshToken godoc
// @Summary      Полное обновление пары токенов (Ротация)
// @Description  Принимает refreshToken, инвалидирует его и возвращает полностью новую пару Access и Refresh токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body     domain.RefreshJwtRequest  true  "Действующий Refresh токен"
// @Success      200     {object}  domain.JwtResponse        "Новая пара токенов"
// @Failure      400     {object}  AuthResponse              "Некорректный формат JSON"
// @Failure      401     {object}  AuthResponse              "Невалидный или просроченный токен"
// @Router       /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.RefreshJwtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMe godoc
// @Summary      Получение информации о текущем пользователе
// @Description  Возвращает профиль пользователя, извлекая UUID из проверенного accessToken в заголовке Authorization.
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Success      200     {object}  UserResponse   "Данные профиля авторизованного пользователя"
// @Failure      401     {object}  AuthResponse   "Токен отсутствует, невалиден или просрочен"
// @Failure      404     {object}  AuthResponse   "Пользователь с таким UUID не найден в БД"
// @Router       /user/me [get]
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuid, ok := h.jwtExt.GetUUID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetById(r.Context(), uuid)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserToResponse(user))
}

func (h *AuthHandler) sendJSON(w http.ResponseWriter, status int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
