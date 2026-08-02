package main

import (
	"jwtAuth/src/internal/di"

	"go.uber.org/fx"
)

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Введите токен в формате: Bearer {accessToken}
func main() {
	fx.New(di.Module).Run()
}
