package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/froggy-12/purpurbase/types"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func IsImage(filename string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".avif", ".jpig"}
	for _, ext := range extensions {
		if filepath.Ext(filename) == ext {
			return true
		}
	}
	return false
}

func IsMusic(filename string) bool {
	extensions := []string{".mp3", ".wav", ".ogg", ".flac", ".m4a"}
	for _, ext := range extensions {
		if filepath.Ext(filename) == ext {
			return true
		}
	}
	return false
}

func IsVideo(filename string) bool {
	ext := filepath.Ext(filename)
	ext = strings.ToLower(ext)
	return ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".wmv"
}

func UploadFile(_ *fiber.Ctx, file *multipart.FileHeader, folder string) (string, error) {
	// Generate a unique filename
	uuid := uuid.New()
	filename := fmt.Sprintf("%s%s", uuid, filepath.Ext(file.Filename))

	// Save the file to the uploads folder
	uploadsDir := filepath.Join(".", "uploads", folder)
	err := os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	filepath := filepath.Join(uploadsDir, filename)
	fileStream, err := file.Open()
	if err != nil {
		return "", err
	}
	defer fileStream.Close()

	// Create the file
	f, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Copy the file contents
	_, err = io.Copy(f, fileStream)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func DeleteFile(filename, folder string) error {
	filepath := filepath.Join(".", "uploads", folder, filename)
	return os.Remove(filepath)
}

func UploadAnyFile(_ *fiber.Ctx, file *multipart.FileHeader, folder, filename string) error {
	// Save the file to the uploads folder
	uploadsDir := filepath.Join(".", "uploads", folder)
	err := os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		return err
	}

	filepath := filepath.Join(uploadsDir, filename)
	fileStream, err := file.Open()
	if err != nil {
		return err
	}
	defer fileStream.Close()

	// Create the file
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Copy the file contents
	_, err = io.Copy(f, fileStream)
	if err != nil {
		return err
	}

	return nil
}

func FindUserFromSQLDBUsingEmail(email string, db *sql.DB) (types.UserSQL, error) {
	var user types.UserSQL
	query := "SELECT ID, UserName, FirstName, LastName, Email, Password, BirthDay, ProfilePicture, CreatedAt, UpdatedAt, Verified, VerificationToken, LastLoggedIn, RawData FROM purpurbase.users WHERE Email = ?;"
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.UserName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.BirthDay,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Verified,
		&user.VerificationToken,
		&user.LastLoggedIn,
		&user.RawData,
	)

	return user, err
}

func FindUserFromSQLDBUsingID(ID string, db *sql.DB) (types.UserSQL, error) {
	var user types.UserSQL
	query := "SELECT ID, UserName, FirstName, LastName, Email, Password, BirthDay, ProfilePicture, CreatedAt, UpdatedAt, Verified, VerificationToken, LastLoggedIn, RawData FROM purpurbase.users WHERE ID = ?;"
	err := db.QueryRow(query, ID).Scan(
		&user.ID,
		&user.UserName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.BirthDay,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Verified,
		&user.VerificationToken,
		&user.LastLoggedIn,
		&user.RawData,
	)

	return user, err
}

func FindUserFromSQLDBUsingUsername(username string, db *sql.DB) (types.UserSQL, error) {
	var user types.UserSQL
	query := "SELECT ID, UserName, FirstName, LastName, Email, Password, BirthDay, ProfilePicture, CreatedAt, UpdatedAt, Verified, VerificationToken, LastLoggedIn, RawData FROM purpurbase.users WHERE UserName = ?;"
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.UserName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.BirthDay,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Verified,
		&user.VerificationToken,
		&user.LastLoggedIn,
		&user.RawData,
	)

	return user, err
}

func LogInSQLDB(c *fiber.Ctx, db *sql.DB, validate validator.Validate, jwtExpirationTime int, jwtSecret string) error {
	var details types.LogInDetails

	if err := c.BodyParser(&details); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
	}

	if err := validate.Struct(&details); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	user, err := FindUserFromSQLDBUsingEmail(details.Email, db)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User does not exist"})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(details.Password))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password"})
	}

	token, err := GenerateJWTToken(user.ID, jwtExpirationTime, jwtSecret)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate JWT token"})
	}

	_, err = db.Exec(`UPDATE purpurbase.users SET LastLoggedIn = CURRENT_TIMESTAMP() WHERE ID = ?`, user.ID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something Went Wrong: " + err.Error()})
	}

	SetJwtHttpCookies(c, token, jwtExpirationTime)
	return c.Status(fiber.StatusAccepted).JSON(types.HTTPSuccessResponse{
		Message: "User has been logged in successfully",
		Data:    map[string]any{"userID": user.ID},
	})
}
