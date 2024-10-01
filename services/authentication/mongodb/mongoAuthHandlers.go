package mongodb

import (
	"context"
	"time"

	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/services/smtpconfigs"
	"github.com/froggy-12/purpurbase/types"
	"github.com/froggy-12/purpurbase/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserWithEmailAndPassword(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
	jwtTokenCookie := c.Cookies("jwtToken")
	if jwtTokenCookie != "" {
		userID, _, _ := utils.ReadJWTToken(jwtTokenCookie, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		if userID != "" {
			return c.Status(fiber.ErrBadRequest.Code).JSON(types.ErrorResponse{Error: "Valid Token Found Please lot out first then try again"})
		}
	}

	collection := mongoClient.Database("purpurbase").Collection("users")

	var user types.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
	}

	if err := validator.Struct(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), config.Configs.PurpurbaseConfigurations.PurpurbasePasswordEncryptionRate)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Failed to Hash the password: " + err.Error()})
	}

	Id := uuid.New().String()
	verificationTokenString := uuid.New().String()

	newUser := types.UserMongo{
		ID:                Id,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		UserName:          user.UserName,
		Email:             user.Email,
		Password:          string(hashedPassword),
		ProfilePicture:    user.ProfilePicture,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		BirthDay:          user.BirthDay,
		Verified:          false,
		VerificationToken: verificationTokenString,
		RawData:           map[string]any{},
	}

	if config.Configs.AuthenticationConfigurations.SetJWTAfterSignUp {
		newUser.LastLoggedIn = time.Now()
	}

	_, err = collection.InsertOne(context.Background(), newUser)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(types.ErrorResponse{Error: "failed to create new user into the database: " + err.Error()})
	}

	if config.Configs.AuthenticationConfigurations.SetJWTAfterSignUp {
		token, err := utils.GenerateJWTToken(newUser.ID, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate jwt token for user: " + newUser.ID})
		}

		utils.SetJwtHttpCookies(c, token, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge)
	}

	if config.Configs.AuthenticationConfigurations.SendEmailAfterSignUpWithToken {
		if !config.Configs.AuthenticationConfigurations.EmailVerification {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email Verification is not configured or turned off please check again and restart the app"})
		}

		if !config.Configs.SMTPConfigurations.SMTPEnabled {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "SMTP is not configured or turned off please check again and restart the app"})
		}

		newToken := uuid.New().String()

		_, err := collection.UpdateOne(context.Background(), bson.M{"id": newUser.ID}, bson.M{"$set": bson.M{"verificationToken": newToken}})

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to update new token: " + err.Error()})
		}

		err = smtpconfigs.SendVerificationEmail(newUser.Email, newToken)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email})
		}

		return c.Status(fiber.StatusAccepted).JSON(types.HTTPSuccessResponse{Message: "User has been created successfully and sent verification email"})

	} else {
		return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{
			Message: "User Has been created to the database hope you will verify the email first then everything",
			Data:    map[string]any{"userID": newUser.ID}})
	}
}

func LogInWithEmailAndPassword(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
	coll := mongoClient.Database("purpurbase").Collection("users")
	token := c.Cookies("jwtToken")

	if token != "" {
		userid, expired, err := utils.ReadJWTToken(token, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		if err != nil || expired {
			err = utils.LogInMongoDB(c, coll, validator, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
			return err
		}

		_, err = utils.FindUserFromMongoDBUsingID(userid, coll)
		if err == nil {
			return c.Status(fiber.StatusAlreadyReported).JSON(types.HTTPSuccessResponse{Message: "You are already logged in"})
		}
	}

	return utils.LogInMongoDB(c, coll, validator, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
}

func SendVerificationEmail(c *fiber.Ctx, mongoClient *mongo.Client) error {
	if !config.Configs.AuthenticationConfigurations.EmailVerification {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email Verification is not configured or turned off please check again and restart the app"})
	}

	if !config.Configs.SMTPConfigurations.SMTPEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "SMTP is not configured or turned off please check again and restart the app"})
	}

	var body struct {
		ID string `json:"id"`
	}

	tokenSet := c.Query("tokenSet", "false")

	if c.Cookies("jwtToken") != "" && tokenSet == "true" {
		userID, _, _ := utils.ReadJWTToken(c.Cookies("jwtToken"), config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		body.ID = userID
	}

	if body.ID == "" && tokenSet == "false" {
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
		}
	}

	coll := mongoClient.Database("purpurbase").Collection("users")

	user, err := utils.FindUserFromMongoDBUsingID(body.ID, coll)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User not found"})
	}

	newToken := uuid.New().String()
	_, err = coll.UpdateOne(context.Background(), bson.M{"id": user.ID}, bson.M{"$set": bson.M{"verificationToken": newToken}})

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to update new token: " + err.Error()})
	}

	err = smtpconfigs.SendVerificationEmail(user.Email, newToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email + " " + err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(types.HTTPSuccessResponse{Message: "Email sent successfully"})

}

func VerifyEmail(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
	coll := mongoClient.Database("purpurbase").Collection("users")

	email := c.Query("email")
	verificationTokenString := c.Query("token")

	if email == "" || verificationTokenString == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email and token are required"})
	}

	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid email: " + err.Error()})
	}

	user, err := utils.FindUserFromMongoDBUsingEmail(email, coll)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(types.ErrorResponse{Error: "User not Found"})
	}

	if user.Verified {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User is already verified"})
	}

	if user.VerificationToken == verificationTokenString {
		_, err := coll.UpdateOne(context.Background(), bson.M{"email": email}, bson.M{"$set": bson.M{"verified": true}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to update user verification status"})
		}
		return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{Message: "Email verified successfully"})
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong token Provided"})
	}

}

func CheckIsEmailAvailable(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
	coll := mongoClient.Database("purpurbase").Collection("users")

	email := c.Query("email")

	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Email: " + err.Error()})
	}
	_, err := utils.FindUserFromMongoDBUsingEmail(email, coll)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{Message: "The Email is good to go"})
}

func CheckIsUsernameAvailable(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("purpurbase").Collection("users")

	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validate.Var(username, "required"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}
	_, err := utils.FindUserFromMongoDBUsingUsername(username, coll)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{Message: "The username is good to go"})
}
