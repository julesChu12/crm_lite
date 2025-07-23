package service

import (
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crypto/tls"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// IEmailService 定义了邮件服务需要实现的接口
type IEmailService interface {
	SendMail(to, subject, body string) error
}

// EmailService 封装了邮件发送的功能
type EmailService struct {
	opts config.EmailOptions
}

// NewEmailService 创建一个新的邮件服务实例
func NewEmailService(opts config.EmailOptions) IEmailService {
	return &EmailService{opts: opts}
}

// SendMail 发送邮件
func (s *EmailService) SendMail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(s.opts.FromAddress, s.opts.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// #nosec G402
	d := gomail.NewDialer(s.opts.Host, s.opts.Port, s.opts.Username, s.opts.Password)
	logger.Info("dialer", zap.String("host", s.opts.Host))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: s.opts.InsecureSkip, ServerName: s.opts.Host}

	return d.DialAndSend(m)
}
