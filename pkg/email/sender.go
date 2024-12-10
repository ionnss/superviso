package email

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailSender struct {
	from     string
	password string
	host     string
	port     string
}

func NewEmailSender() *EmailSender {
	return &EmailSender{
		from:     os.Getenv("EMAIL_ADDRESS"),
		password: os.Getenv("EMAIL_PASSWORD"),
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
	}
}

func (s *EmailSender) SendVerificationEmail(to, token string) error {
	auth := smtp.PlainAuth("", s.from, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	subject := "Verificação de Email - Superviso"
	body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Bem-vindo ao Superviso!</h2>
                <p>Seu código de verificação é: <strong>%s</strong></p>
                <p>Este código expira em 15 minutos.</p>
                <br>
                <p>Se você não solicitou este código, por favor ignore este email.</p>
            </body>
        </html>
    `, token)

	message := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", to, subject, body)

	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
}
