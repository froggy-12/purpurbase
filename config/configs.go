package config

type ExtraConfigurations struct {
	ShowCreditsOnStartup bool `json:"showCreditsOnStartup"`
	DebugLogging         bool `json:"debugLogging"`
}

type DatabaseConfigurations struct {
	DatabaseName            string `json:"databaseName"`
	SQLConnectionURI        string `json:"SQLConnectionURI"`
	MongoDBConnectionURI    string `json:"mongodbConnectionURI"`
	RedisConnectionURI      string `json:"redisConnectionURI"`
	PostgreSQLConnectionURI string `json:"postgreSQLConnectionURI"`
}

type PurpurbaseConfigurations struct {
	PurpurbasePort                   string   `json:"purpurbasePort"`
	PurpurbaseBaseURL                string   `json:"purpurbaseBaseURL"`
	PurpurbaseAllowedCorsOrigins     []string `json:"purpurbaseAllowedCorsOrigins"`
	PurpurbaseAPIServerBodySizeLimit int      `json:"purpurbaseAPIServerBodySizeLimit"`
	PurpurbaseCookieAndCoresAge      int      `json:"purpurbaseCookieAndCoreAge"`
	PurpurbaseJWTTokenSecret         string   `json:"purpurbaseJWTTokenSecret"`
	PurpurbasePasswordEncryptionRate int      `json:"purpurbasePasswordEncryptionRate"`
}

type AuthenticationConfigurations struct {
	Auth                          bool   `json:"auth"`
	OAuth                         bool   `json:"oauth"`
	GoogleOAuth                   bool   `json:"googleOAuth"`
	GoogleOAuthAppID              string `json:"googleOAuthAppID"`
	GoogleOAuthAppSecret          string `json:"googleOAuthAppSecret"`
	GithubOAuth                   bool   `json:"githubOAuth"`
	GithubOAuthAppID              string `json:"githubOAuthAppID"`
	GithubOAuthAppSecret          string `json:"githubOAuthAppSecret"`
	EmailVerification             bool   `json:"emailVerification"`
	SetJWTAfterSignUp             bool   `json:"setJWTAfterSignUp"`
	RealTimeUserData              bool   `json:"realTimeUserData"`
	SendEmailAfterSignUpWithToken bool   `json:"sendEmailAfterSignUpWithToken"`
}

type Features struct {
	ChatFunctionality bool `json:"chatFunctionality"`
	MediaServer       bool `json:"mediaServer"`
	FileUploads       bool `json:"fileUploading"`
}

type SMTPConfigurations struct {
	SMTPEnabled            bool   `json:"smtp_enabled"`              // by default false
	SMTPServerAddress      string `json:"smtp_server_address"`       // by default smtp.gmail.com
	SMTPServerPORT         string `json:"smtp_server_port"`          // by default 587
	SMTPEmailAddrss        string `json:"smtp_email_address"`        // required if SMTPEnabled == true
	SMTPEmailPassword      string `json:"smtp_email_password"`       // required if SMTPEnabled == true
	SMTPAllowedForEveryone bool   `json:"smtp_allowed_for_everyone"` // by default false
}

type Configurations struct {
	PurpurbaseConfigurations     PurpurbaseConfigurations     `json:"purpurbaseConfigurations"`
	AuthenticationConfigurations AuthenticationConfigurations `json:"authConfigurations"`
	ExtraConfigurations          ExtraConfigurations          `json:"configurations"`
	DatabaseConfigurations       DatabaseConfigurations       `json:"databaseConfigs"`
	Features                     Features                     `json:"features"`
	SMTPConfigurations           SMTPConfigurations           `json:"SMTPConfigurations"`
}

var Configs Configurations
