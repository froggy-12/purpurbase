package purpurbasecore

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/froggy-12/purpurbase/api"
	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/database"
	"github.com/froggy-12/purpurbase/internal"
	"github.com/froggy-12/purpurbase/utils"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Purpurbase struct{}

var (
	MongoClient *mongo.Client
	RedisClient *redis.Client
	SQLClient   *sql.DB
)

func Core() *Purpurbase {
	return &Purpurbase{}
}

func (s *Purpurbase) Start() {
	fmt.Println("Checking Configurations please wait.....")
	config.Configs = config.InitConfigs()
	fmt.Println("Global State has been set according to the Configurations")
	fmt.Println("Configurations done going forword now.....")

	if config.Configs.ExtraConfigurations.ShowCreditsOnStartup {
		internal.ShowCredits()
	}

	utils.DebugLogging = config.Configs.ExtraConfigurations.DebugLogging
	utils.DebugLogger("main", "validating some global variebles")

	config.CheckConfigurations()

	utils.DebugLogger("main", "configurations are good to go...")

	utils.DebugLogger("main", "Connecting with databases")

	switch config.Configs.DatabaseConfigurations.DatabaseName {
	case "mongodb":
		MongoClient = database.ConnectToMongoDB(config.Configs.DatabaseConfigurations.MongoDBConnectionURI)
		utils.DebugLogger("main", "Successfully connected with mongodb client")
	case "mysql":
		SQLClient = database.ConnectToSQLClient(config.Configs.DatabaseConfigurations.SQLConnectionURI)
		utils.DebugLogger("main", "Successfully connected with sql client")
	case "postgresql":
		SQLClient = database.ConnectToPostgreSQL(config.Configs.DatabaseConfigurations.PostgreSQLConnectionURI)
		utils.DebugLogger("main", "Successfully connected with sql client")
	default:
		log.Fatal("Unsupported database")
	}

	if config.Configs.Features.ChatFunctionality {
		utils.DebugLogger("main", "found chate functionality is true connecting to redis")
		RedisClient = database.ConnectToRedis(config.Configs.DatabaseConfigurations.RedisConnectionURI)
		utils.DebugLogger("main", "connected with redis")
	}

	utils.DebugLogger("main", "initializing database settings")
	database.Init(MongoClient)

	utils.DebugLogger("main", "Starting the API Server")
	Server := api.NewServer(MongoClient, RedisClient, SQLClient)
	err := Server.StartServer()
	if err != nil {
		log.Fatal("failed to start api server: " + err.Error())
	}
}
