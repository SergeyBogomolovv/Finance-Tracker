package mailer

import (
	"FinanceTracker/notification/internal/config"
	"bytes"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"
)

type mailer struct {
	conf config.SMTP
}

func New(conf config.SMTP) *mailer {
	return &mailer{conf: conf}
}

func (m *mailer) SendOTPEmail(to string, otp string) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", m.conf.User)
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", "Код для входа в Finance Tracker")

	tmpl, err := template.ParseFiles("templates/otp.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{"OTP": otp})
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	mail.SetBody("text/html", body.String())

	d := gomail.NewDialer(m.conf.Host, m.conf.Port, m.conf.User, m.conf.Pass)
	if err := d.DialAndSend(mail); err != nil {
		return fmt.Errorf("failed to send otp email: %w", err)
	}

	return nil
}

func (m *mailer) SendRegistrationEmail(to string) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", m.conf.User)
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", "Регистрация прошла успешно")

	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, nil); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	mail.SetBody("text/html", body.String())

	d := gomail.NewDialer(m.conf.Host, m.conf.Port, m.conf.User, m.conf.Pass)
	if err := d.DialAndSend(mail); err != nil {
		return fmt.Errorf("failed to send register email: %w", err)
	}

	return nil
}
