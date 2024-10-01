package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/froggy-12/purpurbase/types"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var DebugLogging bool = true

func DebugLogger(area, message string) {
	if DebugLogging {
		fmt.Println("[debug]: [" + area + "]   " + message)
	}
}

func GenerateJWTToken(id string, jwtExpirationTime int, jwtSecret string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  id,
		"expr": time.Now().Add(time.Hour * 24 * time.Duration(jwtExpirationTime)).Unix(),
		"iat":  time.Now().Unix(),
	})

	token, err := claims.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", err
	}

	return token, nil
}

func RefreshJWTToken(token, jwtSecret string, jwtExpirationTime int) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	userId, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// Generate a new token with the same user ID and a new expiration time
	newToken, err := GenerateJWTToken(userId, jwtExpirationTime, jwtSecret)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

func ExtractJWTToken(token, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return userID, nil
}

func ReadJWTToken(token, jwtSecret string) (string, bool, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", false, err
	}

	userId, ok := claims["sub"].(string)
	if !ok {
		return "", false, errors.New("invalid token claims")
	}
	expr, ok := claims["expr"].(float64)
	if !ok {
		return "", false, errors.New("invalid token claims")
	}

	expirationTime := time.Unix(int64(expr), 0)
	if time.Now().After(expirationTime) {
		return "", true, nil
	}

	return userId, false, nil
}

func SetJwtHttpCookies(c *fiber.Ctx, token string, cookieAge int) {
	expires := time.Now().Add(time.Hour * 24 * time.Duration(cookieAge))
	maxAge := int(expires.Sub(time.Now()).Seconds())
	cookie := &fiber.Cookie{
		Name:     "jwtToken",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		MaxAge:   maxAge,
	}

	c.Cookie(cookie)
}

func FindUserFromMongoDBUsingEmail(email string, mongoCollection *mongo.Collection) (types.UserMongo, error) {
	filter := bson.M{"email": email}
	user := types.UserMongo{}
	err := mongoCollection.FindOne(context.Background(), filter).Decode(&user)
	return user, err
}

func FindUserFromMongoDBUsingUsername(username string, mongoCollection *mongo.Collection) (types.UserMongo, error) {
	filter := bson.M{"username": username}
	user := types.UserMongo{}
	err := mongoCollection.FindOne(context.Background(), filter).Decode(&user)
	return user, err
}

func FindUserFromMongoDBUsingID(id string, mongoCollection *mongo.Collection) (types.UserMongo, error) {
	filter := bson.M{"id": id}
	user := types.UserMongo{}
	err := mongoCollection.FindOne(context.Background(), filter).Decode(&user)
	return user, err
}

func LogInMongoDB(c *fiber.Ctx, coll *mongo.Collection, validator validator.Validate, jwtTokenTime int, jwtSecret string) error {
	var details types.LogInDetails
	if err := c.BodyParser(&details); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
	}

	if err := validator.Struct(&details); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	user, err := FindUserFromMongoDBUsingEmail(details.Email, coll)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User Doesnt Exist"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(details.Password))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password"})
	}

	token, err := GenerateJWTToken(user.ID, jwtTokenTime, jwtSecret)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate JWT token"})
	}

	lastLoggedIn := time.Now()

	_, err = coll.UpdateOne(context.Background(), bson.M{"id": user.ID}, bson.M{"$set": bson.M{"lastLoggedIn": lastLoggedIn}})

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something Went Wrong while updating last log in informations: " + err.Error()})
	}

	SetJwtHttpCookies(c, token, jwtTokenTime)
	return c.Status(fiber.StatusAccepted).JSON(types.HTTPSuccessResponse{
		Message: "User has been logged in successfully",
		Data:    map[string]any{"userID": user.ID},
	})
}

func LogOut(c *fiber.Ctx) error {
	cookie := &fiber.Cookie{
		Name:     "jwtToken",
		Path:     "/",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		MaxAge:   0,
	}

	c.Cookie(cookie)
	return c.Status(fiber.StatusOK).SendString("User Has been logged out")
}
