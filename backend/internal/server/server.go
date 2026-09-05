package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"btechdevcases/internal/config"
	"btechdevcases/internal/handler"
	"btechdevcases/internal/middleware"
	"btechdevcases/internal/security"
	"btechdevcases/internal/service"
)

type Dependencies struct {
	Config *config.Config
	Pool   *pgxpool.Pool
	Auth   *service.AuthService
	Wallet *service.WalletService
	Tokens *security.JWTManager
}

func New(deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "btech-wallet",
		DisableStartupMessage: deps.Config.Env == "production",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				},
			})
		},
	})

	app.Use(recover.New())
	app.Use(middleware.RequestLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.CORSOrigin,
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Idempotency-Key, X-Request-ID",
		ExposeHeaders:    "X-Access-Token, X-Access-Token-Expires-At, X-Request-ID",
		AllowCredentials: !strings.Contains(deps.Config.CORSOrigin, "*"),
	}))

	authHandler := handler.NewAuthHandler(deps.Auth)
	walletHandler := handler.NewWalletHandler(deps.Wallet)
	healthHandler := handler.NewHealthHandler(deps.Pool)
	protect := middleware.Authenticate(deps.Tokens)

	app.Get("/api/v1/health", healthHandler.Check)

	v1 := app.Group("/api/v1")
	v1.Post("/auth/register", authHandler.Register)
	v1.Post("/auth/login", authHandler.Login)

	authed := v1.Group("", protect)
	authed.Post("/auth/refresh", authHandler.Refresh)
	authed.Post("/auth/logout", authHandler.Logout)
	authed.Get("/me", authHandler.Me)
	authed.Get("/wallet", walletHandler.Get)
	authed.Post("/wallet/transfers", walletHandler.Transfer)

	return app
}
