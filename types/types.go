package types

import (
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
