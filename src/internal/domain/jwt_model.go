package domain

type JwtRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type JwtResponse struct {
	Type         string `json:"type"` // Обычно "Bearer"
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RefreshJwtRequest struct {
	RefreshToken string `json:"refreshToken"`
}
