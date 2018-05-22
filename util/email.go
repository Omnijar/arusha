package util

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"gitlab.com/omnijar/arusha/config"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

const (
	// EnvMailgunDomain for specifying mailgun-activated domain.
	EnvMailgunDomain = "MAILGUN_DOMAIN"
	// EnvMailgunAPIKey for the account.
	EnvMailgunAPIKey = "MAILGUN_API_KEY"
)

var (
	mailgunClient mailgun.Mailgun
	mailgunDomain string
)

// InitializeMailgunClient for sending mails.
func InitializeMailgunClient() error {
	mailgunDomain = os.Getenv(EnvMailgunDomain)
	if mailgunDomain == "" {
		return errors.New(EnvMailgunDomain + " variable not configured or invalid")
	}

	mailgunKey := os.Getenv(EnvMailgunAPIKey)
	if mailgunKey == "" {
		return errors.New(EnvMailgunAPIKey + " variable not configured or invalid")
	}

	mailgunClient = mailgun.NewMailgun(mailgunDomain, mailgunKey, "")
	return nil
}

// SendVerificationMail with the given token (added to the configured URL as query parameter)
// to the the given mail.
func SendVerificationMail(recipient, token string) {
	u, _ := url.Parse(config.Default.EmailVerificationURL)
	q := u.Query()
	q.Add("verify", token)
	u.RawQuery = q.Encode()

	fmt.Println("Verification link:", u.String())

	sender := "postmaster@" + mailgunDomain
	subject := "Please verify your account"
	body := fmt.Sprintf("Hi! To complete your registration process, please click on this link - %s", u.String())

	message := mailgunClient.NewMessage(sender, subject, body, recipient)
	_, _, err := mailgunClient.Send(message)
	if err != nil {
		log.Printf("Error sending mail: %s", err.Error())
	}
}

// SendSecretResetMail with the given token (added to the configured URL as query parameter)
// to the the given mail.
func SendSecretResetMail(recipient, token string) {
	u, _ := url.Parse(config.Default.TokenVerificationURL)
	q := u.Query()
	q.Add("verify", token)
	u.RawQuery = q.Encode()

	fmt.Println("Verification link:", u.String())

	sender := "postmaster@" + mailgunDomain
	subject := "[Password reset]"
	body := fmt.Sprintf("Hi! To reset the password of your account, please click on this link - %s", u.String())

	message := mailgunClient.NewMessage(sender, subject, body, recipient)
	_, _, err := mailgunClient.Send(message)
	if err != nil {
		log.Printf("Error sending mail: %s", err.Error())
	}
}
