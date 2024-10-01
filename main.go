package main

import (
	"github.com/froggy-12/purpurbase/api"
	"github.com/froggy-12/purpurbase/purpurbasecore"
	"github.com/froggy-12/purpurbase/types"
	"github.com/gofiber/fiber/v2"
)

func main() {
	api.GetHandlers = append(api.GetHandlers, types.Handlers{
		Route: "/ping",
		HandlerFunc: func(c *fiber.Ctx) error {
			return c.SendString("pong")
		},
	})

	app := purpurbasecore.Core()
	app.Start()

}
