package service

import (
	"FinanceTracker/notification/internal/config"
	"FinanceTracker/notification/pkg/logger"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"gopkg.in/gomail.v2"
)

const (
	DefaultTimeout = 10 * time.Second
)

type mailService struct {
	conf    config.SMTP
	timeout time.Duration
}

func NewMailService(smtpConf config.SMTP) *mailService {
	return &mailService{
		conf:    smtpConf,
		timeout: DefaultTimeout,
	}
}

func (s *mailService) SendOTP(ctx context.Context, email, code string) error {
	const (
		subject     = "Код для входа в Finance Tracker"
		teplatePath = "templates/otp.html"
	)

	mail := gomail.NewMessage()
	mail.SetHeader("From", s.conf.User)
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", subject)

	tmpl, err := template.ParseFiles(teplatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]string{"Code": code})
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	mail.SetBody("text/html", body.String())

	d := gomail.NewDialer(s.conf.Host, s.conf.Port, s.conf.User, s.conf.Pass)

	errChan := make(chan error, 1)
	go func() {
		errChan <- d.DialAndSend(mail)
	}()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("sending email canceled or timed out: %w", ctx.Err())
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("failed to send otp email: %w", err)
		}
	}

	logger.Debug(ctx, "otp email sent", "email", email)
	return nil
}

func (s *mailService) SendRegistered(ctx context.Context, email, name string) error {
	const (
		subject     = "Добро пожаловать в Finance Tracker"
		teplatePath = "templates/register.html"
	)

	mail := gomail.NewMessage()
	mail.SetHeader("From", s.conf.User)
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", subject)

	tmpl, err := template.ParseFiles(teplatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Name": name}); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	mail.SetBody("text/html", body.String())

	d := gomail.NewDialer(s.conf.Host, s.conf.Port, s.conf.User, s.conf.Pass)

	errChan := make(chan error, 1)
	go func() {
		errChan <- d.DialAndSend(mail)
	}()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("sending email canceled or timed out: %w", ctx.Err())
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("failed to send registration email: %w", err)
		}
	}

	logger.Debug(ctx, "registration email sent", "email", email)
	return nil
}
