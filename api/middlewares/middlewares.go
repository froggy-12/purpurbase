package middlewares

import (
	"strings"
	"time"

	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/types"
	"github.com/froggy-12/purpurbase/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var CorsMiddleWare = cors.New(cors.Config{
	AllowOrigins: strings.Join(config.Configs.PurpurbaseConfigurations.PurpurbaseAllowedCorsOrigins, ", "),
	AllowHeaders: "*",
	AllowMethods: strings.Join([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, ", "),
	MaxAge:       time.Now().Hour() * 24 * config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge,
})

func CheckAndRefreshJWTTokenMiddleware(c *fiber.Ctx) error {
	userId, expired, err := utils.ReadJWTToken(c.Cookies("jwtToken"), config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{Error: "User is not authorised please log in"})
	}
	if expired {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Please Log in"})
	}
	// Pass the token instead of the user ID
	newToken, err := utils.RefreshJWTToken(c.Cookies("jwtToken"), config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
			Error: "Failed to refresh JWT token",
		})
	}

	utils.SetJwtHttpCookies(c, newToken, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge)

	c.Locals("userId", userId)

	return c.Next()
}
