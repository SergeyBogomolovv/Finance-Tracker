package mailer

import (
	"FinanceTracker/notification/internal/config"
	"bytes"
	"fmt"
	"text/template"

	"gopkg.in/gomail.v2"
)

const otpTemplate = `
Здравствуйте!

Ваш OTP-код для входа в аккаунт: {{.OTP}}

Спасибо,
Команда FinanceTracker
`

const registrationTemplate = `
Здравствуйте!

Вы успешно зарегистрировались в FinanceTracker.

Добро пожаловать!

Спасибо,
Команда FinanceTracker
`

type mailer struct {
	conf config.SMTP
}

func New(conf config.SMTP) *mailer {
	return &mailer{conf: conf}
}

func (m *mailer) SendEmail(to, subject, body string) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", m.conf.User)
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/plain", body)

	d := gomail.NewDialer(m.conf.Host, m.conf.Port, m.conf.User, m.conf.Pass)
	if err := d.DialAndSend(mail); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (m *mailer) SendOTPEmail(to string, otp string) error {
	body, err := m.generateBody(otpTemplate, map[string]string{
		"OTP": otp,
	})
	if err != nil {
		return fmt.Errorf("failed to generate email body: %w", err)
	}
	return m.SendEmail(to, "Ваш OTP-код", body)
}

func (m *mailer) SendRegistrationEmail(to string) error {
	body, err := m.generateBody(registrationTemplate, nil)
	if err != nil {
		return fmt.Errorf("failed to generate email body: %w", err)
	}
	return m.SendEmail(to, "Регистрация прошла успешно", body)
}

func (m *mailer) generateBody(tmpl string, data any) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return "", err
	}
	return body.String(), nil
}
