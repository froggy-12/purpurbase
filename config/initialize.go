package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func InitConfigs() Configurations {
	var configs Configurations

	if _, err := os.Stat("configs.json"); os.IsNotExist(err) {
		CreateDefaultConfigurations(&configs)
	} else {
		data, err := os.ReadFile("configs.json")
		if err != nil {
			log.Fatal("Error reading configs.json, error: " + err.Error())
		}
		err = json.Unmarshal(data, &configs)
		if err != nil {
			log.Fatal("Error in convertion of json from configs file, error: " + err.Error())
		}
	}

	return configs
}

func CreateDefaultConfigurations(configs *Configurations) {
	*configs = Configurations{
		PurpurbaseConfigurations: PurpurbaseConfigurations{
			PurpurbasePort:                   "6644",             // default port set 6644
			PurpurbaseBaseURL:                "http://localhost", // default set to localhost
			PurpurbaseAllowedCorsOrigins:     []string{"*"},      // default set to all origin
			PurpurbaseAPIServerBodySizeLimit: 200 * 1024 * 1024,
			PurpurbaseCookieAndCoresAge:      7,
			PurpurbaseJWTTokenSecret:         "SuperSecretPurpurBase",
			PurpurbasePasswordEncryptionRate: 10,
		}, // all default value
		AuthenticationConfigurations: AuthenticationConfigurations{
			Auth:                          true,
			OAuth:                         false,
			GoogleOAuth:                   false,
			GoogleOAuthAppID:              "",
			GoogleOAuthAppSecret:          "",
			GithubOAuth:                   false,
			GithubOAuthAppID:              "",
			GithubOAuthAppSecret:          "",
			EmailVerification:             true,
			SendEmailAfterSignUpWithToken: false,
			SetJWTAfterSignUp:             false,
		},
		ExtraConfigurations: ExtraConfigurations{
			ShowCreditsOnStartup: true,
			DebugLogging:         false,
		},
		DatabaseConfigurations: DatabaseConfigurations{
			DatabaseName: "mongodb",
			SQLClientConfigurations: SQLClientConfigurations{
				Host:     "localhost",
				PORT:     "3306",
				PASSWORD: "purpurbase",
				USER:     "root",
			},
			MongoDBConnectionURI:    "",
			PostgreSQLConnectionURI: "",
			RedisConnectionURI:      "",
		},
		Features: Features{
			ChatFunctionality: true,
			MediaServer:       true,
			FileUploads:       true,
		},
		SMTPConfigurations: SMTPConfigurations{
			SMTPEnabled:            false,
			SMTPServerAddress:      "smtp.gmail.com",
			SMTPServerPORT:         "587",
			SMTPEmailAddrss:        "",
			SMTPEmailPassword:      "",
			SMTPAllowedForEveryone: false,
		},
	}

	data, err := json.MarshalIndent(*configs, "", " ")
	if err != nil {
		log.Fatal("Error creating default configuration file, error: " + err.Error())
	}

	err = os.WriteFile("configs.json", data, 0o644)
	if err != nil {
		log.Fatal("Error writing default configs.json file, error: " + err.Error())
	}

	fmt.Println("Default Configurations has been created restart the app please")
	os.Exit(0)
}

func CheckConfigurations() {
	if Configs.DatabaseConfigurations.DatabaseName != "mongodb" &&
		Configs.DatabaseConfigurations.DatabaseName != "mysql" &&
		Configs.DatabaseConfigurations.DatabaseName != "postgresql" {
		log.Fatal("Wrong Database Name has been provided")
	}
	if Configs.DatabaseConfigurations.DatabaseName == "mongodb" && Configs.DatabaseConfigurations.MongoDBConnectionURI == "" {
		log.Fatal("no uri provided for mongodb")
	} else if Configs.DatabaseConfigurations.DatabaseName == "mysql" && Configs.DatabaseConfigurations.SQLClientConfigurations.Host == "" || Configs.DatabaseConfigurations.SQLClientConfigurations.PASSWORD == "" || Configs.DatabaseConfigurations.SQLClientConfigurations.PORT == "" || Configs.DatabaseConfigurations.SQLClientConfigurations.USER == "" {
		log.Fatal("no uri provided for sql client")
	} else if Configs.DatabaseConfigurations.DatabaseName == "postgresql" && Configs.DatabaseConfigurations.PostgreSQLConnectionURI == "" {
		log.Fatal("no uri provided for postgresql")
	}

	if Configs.Features.ChatFunctionality && Configs.DatabaseConfigurations.RedisConnectionURI == "" {
		log.Fatal("no uri provided for connecting with redis")
	}
}
