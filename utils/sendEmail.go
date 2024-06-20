package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(subject string, html string, to []string) {
	gmailPassword := os.Getenv("GMAILPASSWORD")
	email := os.Getenv("EMAIL")
	auth := smtp.PlainAuth(
		"",
		email,
		gmailPassword,
		"smtp.gmail.com",
	)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"

	msg := "Subject: " + subject + "\n" + headers + "\n\n" + html

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		email,
		to,
		[]byte(msg),
	)

	if err != nil {
		fmt.Println(err)
	}
}
