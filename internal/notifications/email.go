package notifications

import (
	"fmt"
	"net/smtp"
)

type EmailSender struct {
	SMTPServer string
	SMTPPort   int
	Email      string
	Password   string
}

func (e *EmailSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", e.SMTPServer, e.SMTPPort)

	msg := []byte(
		"From: " + e.Email + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n\r\n" +
			body,
	)

	auth := smtp.PlainAuth("", e.Email, e.Password, e.SMTPServer)

	return smtp.SendMail(addr, auth, e.Email, []string{to}, msg)
}
