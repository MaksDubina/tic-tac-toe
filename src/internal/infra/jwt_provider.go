package infra

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"jwtAuth/src/internal/config"
	"jwtAuth/src/internal/domain"
	"strings"
	"time"
)

type TokenClaims struct {
	Sub  string `json:"sub"`
	ID   string `json:"id"`
	Exp  int64  `json:"exp"`
	Type string `json:"type"`
}

type JwtProvider struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewJwtProvider(cfg *config.Config) *JwtProvider {
	return &JwtProvider{
		accessSecret:  []byte(cfg.JWTAccessSecret),
		refreshSecret: []byte(cfg.JWTRefreshSecret),
	}
}

func (jp *JwtProvider) GenerateAccessToken(user domain.User) (string, error) {
	exp := time.Now().Add(15 * time.Minute).Unix()
	return jp.createToken(user.Login, user.ID.String(), "access", exp, jp.accessSecret)
}

func (jp *JwtProvider) GenerateRefreshToken(user domain.User) (string, error) {
	exp := time.Now().Add(30 * 24 * time.Hour).Unix()
	return jp.createToken(user.Login, user.ID.String(), "refresh", exp, jp.refreshSecret)
}

func (jp *JwtProvider) ValidateAccessToken(tokenStr string) bool {
	claims, err := jp.parseAndValidate(tokenStr, jp.accessSecret)
	return err == nil && claims.Type == "access"
}

func (jp *JwtProvider) ValidateRefreshToken(tokenStr string) bool {
	claims, err := jp.parseAndValidate(tokenStr, jp.refreshSecret)
	return err == nil && claims.Type == "refresh"
}

func (jp *JwtProvider) GetIdFromToken(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var claims TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", err
	}
	return claims.ID, nil
}

func (jp *JwtProvider) createToken(sub, ID, tType string, exp int64, secret []byte) (string, error) {
	header := `{"alg":"HS256","typ":"JWT"}`
	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(header))

	claims := TokenClaims{Sub: sub, ID: ID, Exp: exp, Type: tType}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	unsignedToken := encodedHeader + "." + encodedPayload
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(unsignedToken))
	encodedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return unsignedToken + "." + encodedSignature, nil
}

func (jp *JwtProvider) parseAndValidate(tokenStr string, secret []byte) (*TokenClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token")
	}

	unsignedToken := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(unsignedToken))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, fmt.Errorf("signature mismatch")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, err
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}
