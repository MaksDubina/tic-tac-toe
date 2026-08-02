package api

type contextKey string

const userIDKey contextKey = "user_id"

type SignUpRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success,omitempty"`
	UserID  string `json:"user_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UserResponse struct {
	ID    string `json:"user_id,omitempty"`
	Login string `json:"login"`
}
