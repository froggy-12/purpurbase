package types

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Route       string
	HandlerFunc fiber.Handler
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SingleFileUploadedSuccessResponse struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}

type MultipleFileUploadedSuccessResponse struct {
	FileNames []string `json:"fileNames"`
	Message   string   `json:"message"`
}

type DeleteSuccessResponse struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}

type HTTPSuccessResponse struct {
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

type User struct {
	ID             string    `json:"id"`
	UserName       string    `json:"username" validate:"required"`
	FirstName      string    `json:"firstName" validate:"required"`
	LastName       string    `json:"lastName" validate:"required"`
	Email          string    `json:"email" validate:"required,email"`
	BirthDay       time.Time `json:"birthday"`
	Password       string    `json:"password" validate:"required,min=8"`
	ProfilePicture string    `json:"profilePicture"`
}

type UserMongo struct {
	ID                string         `bson:"id"`
	UserName          string         `bson:"username, unique"`
	FirstName         string         `bson:"firstName"`
	LastName          string         `bson:"lastName"`
	Email             string         `bson:"email, unique"`
	Password          string         `bson:"password"`
	BirthDay          time.Time      `bson:"birthday"`
	ProfilePicture    string         `bson:"profilePicture"`
	CreatedAt         time.Time      `bson:"createdAt"`
	UpdatedAt         time.Time      `bson:"updatedAt"`
	Verified          bool           `bson:"verified"`
	VerificationToken string         `bson:"verificationToken"`
	LastLoggedIn      time.Time      `bson:"lastLoggedIn"`
	RawData           map[string]any `bson:"rawData"`
}

type LogInDetails struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserSQL struct {
	ID                string          `json:"id"`
	UserName          string          `json:"username"`
	FirstName         string          `json:"firstName"`
	LastName          string          `json:"lastName"`
	Email             string          `json:"email"`
	Password          string          `json:"password"`
	BirthDay          time.Time       `json:"birthDay"`
	ProfilePicture    string          `json:"profilePicture"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	Verified          bool            `json:"verified"`
	VerificationToken string          `json:"verificationToken"`
	LastLoggedIn      time.Time       `json:"lastLoggedIn"`
	RawData           json.RawMessage `json:"rawData"`
}
