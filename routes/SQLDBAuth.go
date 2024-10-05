package routes

import (
	"database/sql"

	"github.com/froggy-12/purpurbase/services/authentication/sqldb"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func SQLDBAuth(router fiber.Router, SQLClient *sql.DB) {
	validator := validator.New()
	router.Post("/create-user", func(c *fiber.Ctx) error {
		return sqldb.CreateUserWithEmailAndPassword(c, SQLClient, *validator)
	})
	router.Post("/log-in", func(c *fiber.Ctx) error {
		return sqldb.LogInWithEmailAndPassword(c, SQLClient, *validator)
	})
	router.Post("/send-verification-email", func(c *fiber.Ctx) error {
		return sqldb.SendVerificationEmail(c, SQLClient, *validator)
	})
	router.Get("/verified", func(c *fiber.Ctx) error {
		return sqldb.VerifyEmail(c, SQLClient, *validator)
	})
	router.Get("/check-email-availability", func(c *fiber.Ctx) error {
		return sqldb.CheckIsEmailAvailable(c, SQLClient, *validator)
	})
	router.Get("/check-username-availability", func(c *fiber.Ctx) error {
		return sqldb.CheckIsUsernameAvailable(c, SQLClient, *validator)
	})
}
