package routes

import (
	"github.com/froggy-12/purpurbase/services/authentication/mongodb"
	"github.com/froggy-12/purpurbase/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func MongoUserRoutes(router fiber.Router, mongoClient *mongo.Client) {
	validator := validator.New()
	router.Get("/get-user", func(c *fiber.Ctx) error {
		return mongodb.GetUser(c, mongoClient)
	})
	router.Put("/update-username", func(c *fiber.Ctx) error {
		return mongodb.UpdateUserName(c, mongoClient, *validator)
	})
	router.Put("/update-user-info", func(c *fiber.Ctx) error {
		return mongodb.UpdateUser(c, mongoClient)
	})
	router.Put("/update-email", func(c *fiber.Ctx) error {
		return mongodb.ChangeEmail(c, mongoClient, *validator)
	})
	router.Put("/append-raw-data", func(c *fiber.Ctx) error {
		return mongodb.AddRawData(c, mongoClient)
	})
	router.Put("/change-password", func(c *fiber.Ctx) error {
		return mongodb.ChangePassword(c, mongoClient, *validator)
	})
	router.Delete("/delete-user", func(c *fiber.Ctx) error {
		return mongodb.DeleteUser(c, mongoClient, *validator)
	})
	router.Get("/log-out", func(c *fiber.Ctx) error {
		return utils.LogOut(c)
	})
}
