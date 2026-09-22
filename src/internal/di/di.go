package di

import (
	"context"
	"database/sql"
	"embed"
	"jwtAuth/src/internal/api"
	"jwtAuth/src/internal/app"
	"jwtAuth/src/internal/config"
	"jwtAuth/src/internal/infra"
	"jwtAuth/src/internal/metrics"
	"net/http"

	_ "jwtAuth/src/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/fx"
)

//go:embed migrations
var embedMigrations embed.FS

func RunMigrations(cfg *config.Config) error {

	db, err := sql.Open("pgx", cfg.GetDSN())
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	println("[Goose] Миграции успешно применены!")
	return nil
}

var Module = fx.Options(
	fx.Provide(
		config.NewConfig,
		infra.NewPostgresPool,

		fx.Annotate(
			infra.NewUserRepository,
			fx.As(new(app.UserRepository)),
		),

		fx.Annotate(
			infra.NewRepository,
			fx.As(new(app.GameRepository)),
		),

		fx.Annotate(
			infra.NewJwtProvider,
			fx.As(new(app.TokenProvider)),
		),
		fx.Annotate(
			infra.NewJwtProvider,
			fx.As(new(api.TokenProvider)),
		),

		fx.Annotate(
			app.NewService,
			fx.As(new(api.GameService)),
		),
		fx.Annotate(
			app.NewAuthService,
			fx.As(new(api.AuthService)),
		),

		api.NewGameHandler,
		api.NewAuthHandler,
		api.NewUserAuthenticator,
	),
	fx.Provide(metrics.New),
	fx.Invoke(RunMigrations, StartHTTPServer),
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func StartHTTPServer(lc fx.Lifecycle, gameHandler *api.GameHandler, authHandler *api.AuthHandler,
	authenticator *api.UserAuthenticator, pool *pgxpool.Pool, m *metrics.Metrics) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/signup", authHandler.HandleRegister)
	mux.HandleFunc("POST /auth/signin", authHandler.HandleLogin)
	mux.HandleFunc("POST /auth/refresh-access", authHandler.RefreshAccess)
	mux.HandleFunc("POST /auth/refresh-token", authenticator.Wrap(authHandler.RefreshToken))

	mux.HandleFunc("POST /game", authenticator.Wrap(gameHandler.CreateGame))
	mux.HandleFunc("POST /game/{current_game_UUID}", authenticator.Wrap(gameHandler.HandleMove))
	mux.HandleFunc("POST /game/{current_game_UUID}/join", authenticator.Wrap(gameHandler.JoinGame))
	mux.HandleFunc("GET /game/{current_game_UUID}", authenticator.Wrap(gameHandler.GetGame))
	mux.HandleFunc("GET /games/available", authenticator.Wrap(gameHandler.GetAvailableGames))
	mux.HandleFunc("GET /games/history", authenticator.Wrap(gameHandler.GetGamesHistory))
	mux.HandleFunc("GET /games/leaderboard", authenticator.Wrap(gameHandler.GetLiederBoard))

	mux.HandleFunc("GET /user/me", authenticator.Wrap(authHandler.GetMe))
	mux.HandleFunc("GET /users/{user_uuid}", authenticator.Wrap(gameHandler.GetUser))

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    ":8080",
		Handler: enableCORS(api.MetricsMiddleware(m)(mux)),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				println("[HTTP] Сервер запущен на порту :8080")
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					println("[HTTP] Критическая ошибка сервера:", err.Error())
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			println("[SHUTDOWN] Начинаем плавное завершение HTTP-сервера...")

			if err := srv.Shutdown(ctx); err != nil {
				println("[SHUTDOWN] Ошибка остановки сервера:", err.Error())
				return err
			}
			println("[SHUTDOWN] HTTP-сервер успешно остановлен.")

			println("[SHUTDOWN] Закрываем пул соединений PostgreSQL...")
			pool.Close()
			println("[SHUTDOWN] Пул БД успешно закрыт.")

			return nil
		},
	})

	return srv
}
