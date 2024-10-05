package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init(mongoClient *mongo.Client, SQLDB *sql.DB) {
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

	if config.Configs.DatabaseConfigurations.DatabaseName == "mysql" {
		utils.DebugLogger("db", "detected mariadb as primary database running some configurations")

		_, err := SQLDB.Exec(`CREATE DATABASE IF NOT EXISTS purpurbase;`)

		if err != nil {
			log.Fatal(err)
		}

		_, err = SQLDB.Exec(`
			CREATE TABLE IF NOT EXISTS purpurbase.users (
				ID VARCHAR(255) NOT NULL,
				UserName VARCHAR(255) NOT NULL UNIQUE,
				FirstName VARCHAR(255) NOT NULL,
				LastName VARCHAR(255) NOT NULL,
				Email VARCHAR(255) NOT NULL UNIQUE,
				Password VARCHAR(255) NOT NULL,
				BirthDay DATE,
				ProfilePicture VARCHAR(255),
				CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				UpdatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				Verified BOOLEAN NOT NULL DEFAULT FALSE,
				VerificationToken VARCHAR(255),
				LastLoggedIn TIMESTAMP,
				RawData JSON,
				PRIMARY KEY (ID)
			);
		`)

		if err != nil {
			log.Fatal(err)
		}

	}

}
