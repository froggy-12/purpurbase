package database

import (
	"context"
	"log"

	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init(mongoClient *mongo.Client) {
	if config.Configs.DatabaseConfigurations.DatabaseName == "mongodb" {
		utils.DebugLogger("db", "detected mongodb as primary database indexing and checking some models")
		database := mongoClient.Database("purpurbase")
		usersCollection := database.Collection("users")

		_, err := usersCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.M{"email": 1},
			Options: options.Index().SetUnique(true),
		})

		if err != nil {
			log.Fatal(err)
		}

		_, err = usersCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.M{"username": 1},
			Options: options.Index().SetUnique(true),
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}
