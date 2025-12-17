package notifications

import (
	"encoding/json"
	"fmt"
	"os"
)

type EmailConfig struct {
	SMTPServer string `json:"smtp_server"`
	SMTPPort   int    `json:"smtp_port"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Recipient  string `json:"recipient"`
}

var sender *EmailSender
var recipient string
var loaded bool

func loadEmailConfig() error {
	if loaded {
		return nil
	}

	data, err := os.ReadFile("config/email.json")
	if err != nil {
		return err
	}

	var cfg EmailConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	sender = &EmailSender{
		SMTPServer: cfg.SMTPServer,
		SMTPPort:   cfg.SMTPPort,
		Email:      cfg.Email,
		Password:   cfg.Password,
	}
	recipient = cfg.Recipient
	loaded = true

	return nil
}

func SendEmailNotification(url, message string) error {
	if err := loadEmailConfig(); err != nil {
		return fmt.Errorf("email config error: %w", err)
	}

	subject := "Website Change Detected"
	body := fmt.Sprintf("URL: %s\n\n%s", url, message)

	return sender.Send(recipient, subject, body)
}
