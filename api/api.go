package api

import (
	"database/sql"

	"github.com/froggy-12/purpurbase/api/middlewares"
	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/routes"
	"github.com/froggy-12/purpurbase/services/mediaserver"
	"github.com/froggy-12/purpurbase/types"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	mongoClient *mongo.Client
	redisClient *redis.Client
	sqlClient   *sql.DB
}

var (
	GetHandlers    []types.Handlers
	PostHandlers   []types.Handlers
	PutHandlers    []types.Handlers
	PatchHandlers  []types.Handlers
	DeleteHandlers []types.Handlers
	HeadHandlers   []types.Handlers
	OptionHandlers []types.Handlers
)

func NewServer(mongoClient *mongo.Client, redisClient *redis.Client, sqlClient *sql.DB) *Server {
	return &Server{
		mongoClient: mongoClient,
		redisClient: redisClient,
		sqlClient:   sqlClient,
	}
}

func (s *Server) StartServer() error {
	app := fiber.New(fiber.Config{
		BodyLimit:       config.Configs.PurpurbaseConfigurations.PurpurbaseAPIServerBodySizeLimit,
		ServerHeader:    "HTTPS",
		Concurrency:     256 * 1024,
		ReadBufferSize:  8192,
		WriteBufferSize: 8192,
	})

	if config.Configs.ExtraConfigurations.DebugLogging {
		app.Use(logger.New())
	}

	for _, handler := range GetHandlers {
		app.Get(handler.Route, handler.HandlerFunc)
	}
	for _, handler := range PostHandlers {
		app.Post(handler.Route, handler.HandlerFunc)
	}
	for _, handler := range PutHandlers {
		app.Put(handler.Route, handler.HandlerFunc)
	}
	for _, handler := range PatchHandlers {
		app.Patch(handler.Route, handler.HandlerFunc)
	}
	for _, handler := range OptionHandlers {
		app.Options(handler.Route, handler.HandlerFunc)
	}
	for _, handler := range DeleteHandlers {
		app.Delete(handler.Route, handler.HandlerFunc)
	}

	freeRouter := app.Group("/api", middlewares.CorsMiddleWare)
	routes.FreeRoutes(freeRouter)

	if config.Configs.Features.FileUploads {
		routes.FileUploadingRoutes(freeRouter)
	}

	if config.Configs.Features.MediaServer {
		freeRouter.Get("/get_file", mediaserver.ServeFiles)
		freeRouter.Get("/download_file", mediaserver.DownloadFile)
	}

	if config.Configs.AuthenticationConfigurations.Auth {
		if config.Configs.DatabaseConfigurations.DatabaseName == "mongodb" {
			authRouter := app.Group("/api/auth")
			userRouter := app.Group("/api/data", middlewares.CheckAndRefreshJWTTokenMiddleware)
			routes.MongoAuthRoutes(authRouter, s.mongoClient)
			routes.MongoUserRoutes(userRouter, s.mongoClient)
		}
	}

	return app.Listen(":" + config.Configs.PurpurbaseConfigurations.PurpurbasePort)
}
