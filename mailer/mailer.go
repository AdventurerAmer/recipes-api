package mailer

import (
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	from   string
	dialer *gomail.Dialer
}

func New(cfg *config.Mailer) *Mailer {
	d := gomail.NewDialer(cfg.Host, *cfg.Port, cfg.Username, cfg.Password)
	return &Mailer{
		from:   cfg.From,
		dialer: d,
	}
}

type Message struct {
	To      string
	Subject string
	Body    string
}

func (m *Mailer) Send(mail Message) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", mail.To)
	msg.SetHeader("Subject", mail.Subject)
	msg.SetBody("text/html", mail.Body)

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("'dialer.DialAndSend' failed: %w", err)
	}

	return nil
}
