package sqldb

import (
	"database/sql"

	"github.com/froggy-12/purpurbase/config"
	"github.com/froggy-12/purpurbase/services/smtpconfigs"
	"github.com/froggy-12/purpurbase/types"
	"github.com/froggy-12/purpurbase/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserWithEmailAndPassword(c *fiber.Ctx, sqlClient *sql.DB, validate validator.Validate) error {
	jwtTokenCookie := c.Cookies("jwtToken")

	if jwtTokenCookie != "" {
		userID, _, _ := utils.ReadJWTToken(jwtTokenCookie, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		if userID != "" {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Valid Token found please log out first then sign up"})
		}
	}

	var user types.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
	}

	if err := validate.Struct(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), config.Configs.PurpurbaseConfigurations.PurpurbasePasswordEncryptionRate)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Failed to Hash the password: " + err.Error()})
	}

	Id := uuid.New().String()

	verificationTokenString := uuid.New().String()

	newUser := types.UserSQL{
		ID:                Id,
		UserName:          user.UserName,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Email:             user.Email,
		Password:          string(hashedPassword),
		BirthDay:          user.BirthDay,
		ProfilePicture:    user.ProfilePicture,
		Verified:          false,
		VerificationToken: verificationTokenString,
	}

	_, err = sqlClient.Exec(`
	INSERT INTO purpurbase.users (
		ID,
		UserName,
		FirstName,
		LastName,
		Email,
		Password,
		BirthDay,
		ProfilePicture,
		Verified,
		VerificationToken,
		LastLoggedIn,
		RawData
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		Id,
		newUser.UserName,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		newUser.Password,
		newUser.BirthDay,
		newUser.ProfilePicture,
		newUser.Verified,
		newUser.VerificationToken,
		nil, // LastLoggedIn
		`{}`,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to create user: " + err.Error()})
	}

	if config.Configs.AuthenticationConfigurations.SetJWTAfterSignUp {
		token, err := utils.GenerateJWTToken(newUser.ID, 1, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate jwt token for user: " + newUser.ID})
		}
		utils.SetJwtHttpCookies(c, token, 1)
	}

	if config.Configs.AuthenticationConfigurations.SendEmailAfterSignUpWithToken {
		if !config.Configs.AuthenticationConfigurations.EmailVerification {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email Verification is not configured or turned off please check again and restart the app"})
		}

		if !config.Configs.SMTPConfigurations.SMTPEnabled {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "SMTP is not configured or turned off please check again and restart the app"})
		}

		err = smtpconfigs.SendVerificationEmail(newUser.Email, newUser.VerificationToken)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email})
		}

		return c.Status(fiber.StatusAccepted).JSON(types.HTTPSuccessResponse{Message: "User has been created successfully and sent verification email"})

	}

	return c.Status(fiber.StatusCreated).JSON(types.HTTPSuccessResponse{
		Message: "User Has been created to the database hope you will verify the email first then everything",
		Data:    map[string]any{"userId": newUser.ID},
	})

}

func LogInWithEmailAndPassword(c *fiber.Ctx, mariadbClient *sql.DB, validate validator.Validate) error {
	token := c.Cookies("jwtToken")
	if token != "" {
		userid, expired, err := utils.ReadJWTToken(token, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		if err != nil || expired {
			return utils.LogInSQLDB(c, mariadbClient, validate, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
		}

		_, err = utils.FindUserFromSQLDBUsingID(userid, mariadbClient)
		if err == nil {
			return c.Status(fiber.StatusAlreadyReported).JSON(types.HTTPSuccessResponse{Message: "You are already logged in"})
		}
	}

	return utils.LogInSQLDB(c, mariadbClient, validate, config.Configs.PurpurbaseConfigurations.PurpurbaseCookieAndCoresAge, config.Configs.PurpurbaseConfigurations.PurpurbaseJWTTokenSecret)
}

func SendVerificationEmail(c *fiber.Ctx, mariadbClient *sql.DB, validate validator.Validate) error {
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

	if body.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "No ID Has been found"})
	}

	user, err := utils.FindUserFromSQLDBUsingID(body.ID, mariadbClient)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User not found"})
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
		}
	}

	if user.Verified {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User is already verified"})
	}

	newToken := uuid.New().String()

	_, err = mariadbClient.Exec(`UPDATE mooshroombase.users SET verificationToken = ? WHERE ID = ?`, newToken, user.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "failed to generate and set new verification token: " + err.Error()})
	}

	err = smtpconfigs.SendVerificationEmail(user.Email, newToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email})
	}

	return c.Status(fiber.StatusAccepted).JSON(types.HTTPSuccessResponse{Message: "Email sent successfully"})
}

func VerifyEmail(c *fiber.Ctx, mariadbClient *sql.DB, validator validator.Validate) error {
	email := c.Query("email")
	verificationTokenString := c.Query("token")

	if email == "" || verificationTokenString == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email and token are required"})
	}

	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid email: " + err.Error()})
	}

	user, err := utils.FindUserFromSQLDBUsingEmail(email, mariadbClient)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusBadGateway).JSON(types.ErrorResponse{Error: "User not Found"})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
		}
	}

	if user.Verified {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User is already verified"})
	}

	if user.VerificationToken == verificationTokenString {
		_, err := mariadbClient.Exec(`UPDATE purpurbase.users SET Verified = true WHERE ID = ?`, user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrorResponse{Error: "failed to update user's verification status: " + err.Error()})
		}
		return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{Message: "Email verified successfully"})
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong token Provided"})
	}
}

func CheckIsEmailAvailable(c *fiber.Ctx, mariaDBClient *sql.DB, validator validator.Validate) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Email: " + err.Error()})
	}
	_, err := utils.FindUserFromSQLDBUsingEmail(email, mariaDBClient)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{Message: "The Email is good to go"})
}

func CheckIsUsernameAvailable(c *fiber.Ctx, mariaDBClient *sql.DB, validator validator.Validate) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(username, "required"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}
	_, err := utils.FindUserFromSQLDBUsingUsername(username, mariaDBClient)
	if err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(fiber.StatusOK).JSON(types.HTTPSuccessResponse{Message: "The username is good to go"})
}
