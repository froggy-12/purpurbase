package routes

import (
	"github.com/froggy-12/purpurbase/services/authentication/mongodb"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func MongoAuthRoutes(router fiber.Router, mongoClient *mongo.Client) {

	validator := validator.New()

	router.Post("/create-user", func(c *fiber.Ctx) error {
		return mongodb.CreateUserWithEmailAndPassword(c, mongoClient, *validator)
	})

	router.Post("/log-in", func(c *fiber.Ctx) error {
		return mongodb.LogInWithEmailAndPassword(c, mongoClient, *validator)
	})

	router.Post("/send-verification-email", func(c *fiber.Ctx) error {
		return mongodb.SendVerificationEmail(c, mongoClient)
	})

	router.Get("/verified", func(c *fiber.Ctx) error {
		return mongodb.VerifyEmail(c, mongoClient, *validator)
	})

	router.Get("/check-email-availability", func(c *fiber.Ctx) error {
		return mongodb.CheckIsEmailAvailable(c, mongoClient, *validator)
	})
	router.Get("/check-username-availability", func(c *fiber.Ctx) error {
		return mongodb.CheckIsUsernameAvailable(c, mongoClient, *validator)
	})
}
