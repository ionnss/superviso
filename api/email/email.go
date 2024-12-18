package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	From     string
	Password string
	Host     string
	Port     string
}

func getConfig() EmailConfig {
	return EmailConfig{
		From:     os.Getenv("EMAIL_ADDRESS"),
		Password: os.Getenv("EMAIL_PASSWORD"),
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
	}
}

func SendVerificationEmail(to, token string) error {
	config := getConfig()
	auth := smtp.PlainAuth("", config.From, config.Password, config.Host)

	templateData := struct {
		VerificationLink string
	}{
		VerificationLink: fmt.Sprintf("http://localhost:8080/verify-email?token=%s", token),
	}

	// Ler o template
	tmpl, err := template.ParseFiles("view/email_templates/verification.html")
	if err != nil {
		return fmt.Errorf("erro ao carregar template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, templateData); err != nil {
		return fmt.Errorf("erro ao executar template: %v", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: Confirme seu email - Superviso\n"
	msg := []byte(subject + mime + body.String())

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	if err := smtp.SendMail(addr, auth, config.From, []string{to}, msg); err != nil {
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	return nil
}

func SendEmail(to, subject, body string) error {
	config := getConfig()
	auth := smtp.PlainAuth("", config.From, config.Password, config.Host)

	// Ler a imagem do logo
	logoPath := "static/assets/img/logo.png"
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		return fmt.Errorf("erro ao ler logo: %v", err)
	}

	// Criar email com múltiplas partes
	boundary := "superviso-boundary"
	mime := fmt.Sprintf(
		"MIME-Version: 1.0\n"+
			"Content-Type: multipart/related; boundary=%s\n\n"+
			"--%s\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\n\n"+
			"%s\n\n"+
			"--%s\n"+
			"Content-Type: image/png\n"+
			"Content-ID: <logo>\n"+
			"Content-Transfer-Encoding: base64\n\n"+
			"%s\n"+
			"--%s--",
		boundary, boundary, body, boundary,
		base64.StdEncoding.EncodeToString(logo), boundary,
	)

	msg := []byte(fmt.Sprintf("Subject: %s\n%s", subject, mime))
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	return smtp.SendMail(addr, auth, config.From, []string{to}, msg)
}
