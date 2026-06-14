package service

import (
    "crypto/tls"
    "fmt"
    "log"

    "bank-api/config"
    "github.com/go-mail/mail/v2"
)

type EmailService struct {
    dialer *mail.Dialer
    from   string
}

func NewEmailService(cfg *config.Config) *EmailService {
    d := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
    d.TLSConfig = &tls.Config{InsecureSkipVerify: false, ServerName: cfg.SMTPHost}
    return &EmailService{
        dialer: d,
        from:   cfg.SMTPUser,
    }
}

func (e *EmailService) SendPaymentNotification(to string, amount float64, purpose string) error {
    subject := "Банковское уведомление"
    body := fmt.Sprintf(`
        <h2>Произошёл платёж</h2>
        <p>Сумма: <strong>%.2f RUB</strong></p>
        <p>Назначение: %s</p>
        <small>Автоматическое уведомление</small>
    `, amount, purpose)

    m := mail.NewMessage()
    m.SetHeader("From", e.from)
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)

    if err := e.dialer.DialAndSend(m); err != nil {
        log.Printf("SMTP error: %v", err)
        return err
    }
    log.Printf("Email sent to %s", to)
    return nil
}
