package smtpconfigs

import (
	"bytes"
	"html/template"
	"net/smtp"

	"github.com/froggy-12/purpurbase/config"
)

func SendVerificationEmail(emailTo, token string) error {
	type EmailData struct {
		Token string
	}

	tmpl := template.Must(template.New("email").Parse(`<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Lets get your verified 🐱</title>
  <style>
    * {
      font-family: "Arial", sans-serif;
      font-weight: bold;
    }

    body {
      background-color: rgb(17 24 39);
      color: white;
      border-radius: 20px;
      padding: 40px;
    }
    .container {
      padding: 20px;
      display: flex;
      justify-content: center;
      align-items: center;
      flex-direction: column;
    }
  </style>
</head>

<body>
  <div class="container">
    <h1>
      Your verification token is:
    </h1>
    <p>{{ .Token }}</p>
  </div>
</body>

</html>`))

	var buffer bytes.Buffer
	data := EmailData{Token: token}    // Create an instance of EmailData
	err := tmpl.Execute(&buffer, data) // Pass the data instance to tmpl.Execute
	if err != nil {
		return err
	}

	msg := "To: " + emailTo + "\r\n" +
		"Subject: " + "Email verification (Purpurbase)" + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		buffer.String()

	auth := smtp.PlainAuth(config.Configs.SMTPConfigurations.SMTPEmailAddrss, config.Configs.SMTPConfigurations.SMTPEmailAddrss, config.Configs.SMTPConfigurations.SMTPEmailPassword, config.Configs.SMTPConfigurations.SMTPServerAddress)
	err = smtp.SendMail(config.Configs.SMTPConfigurations.SMTPServerAddress+":"+config.Configs.SMTPConfigurations.SMTPServerPORT, auth, config.Configs.SMTPConfigurations.SMTPEmailAddrss, []string{emailTo}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
