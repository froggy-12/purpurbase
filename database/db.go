package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/froggy-12/purpurbase/config"
	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB(connectionURI string) *mongo.Client {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectionURI))
	if err != nil {
		log.Fatal("Failed to connect with MongoDB 🍃: " + err.Error())
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Failed to Ping MongoDB: " + err.Error())
	}

	return client
}

func ConnectToRedis(connectionURI string) *redis.Client {
	opt, err := redis.ParseURL(connectionURI)
	if err != nil {
		log.Fatal("Failed to format redis connection uri: " + err.Error())
	}

	client := redis.NewClient(opt)
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect with redis 🔴: " + err.Error())
	}

	return client
}

func ConnectToSQLClient() *sql.DB {
	cfg := mysql.Config{
		User:                 config.Configs.DatabaseConfigurations.SQLClientConfigurations.USER,
		Passwd:               config.Configs.DatabaseConfigurations.SQLClientConfigurations.PASSWORD,
		Net:                  "tcp",
		Addr:                 config.Configs.DatabaseConfigurations.SQLClientConfigurations.Host + ":" + config.Configs.DatabaseConfigurations.SQLClientConfigurations.PORT,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("Failed to create sql client instance: " + err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error Connecting to SQL Client 🐬: ", err.Error())
	}

	return db
}

func ConnectToPostgreSQL(connectionURI string) *sql.DB {
	db, err := sql.Open("postgres", connectionURI)
	if err != nil {
		log.Fatal("Failed to create postgres client instance: " + err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error Connecting to PostgreSQL 🐬: ", err.Error())
	}

	return db
}
