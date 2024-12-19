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

type EmailTemplateData struct {
	Title      string
	Message    template.HTML
	LogoBase64 string
	Date       string
	Time       string
	ActionURL  string
	ActionText string
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

	// Ler a imagem do logo
	logoPath := "static/assets/email/logo.png"
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		return fmt.Errorf("erro ao ler logo: %v", err)
	}

	// Converter logo para base64
	logoBase64 := base64.StdEncoding.EncodeToString(logo)

	templateData := struct {
		VerificationLink string
		LogoBase64       string
	}{
		VerificationLink: fmt.Sprintf("http://localhost:8080/verify-email?token=%s", token),
		LogoBase64:       logoBase64,
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

	mime := "MIME-Version: 1.0\n" +
		"Content-Type: text/html; charset=UTF-8\n\n"
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
	logoPath := "static/assets/email/logo.png"
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		return fmt.Errorf("erro ao ler logo: %v", err)
	}

	// Converter logo para base64
	logoBase64 := base64.StdEncoding.EncodeToString(logo)

	templateData := EmailTemplateData{
		Title:      subject,
		Message:    template.HTML(body),
		LogoBase64: logoBase64,
		ActionURL:  "",
		ActionText: "",
	}

	// Carregar e executar template
	tmpl, err := template.ParseFiles("view/email_templates/notification.html")
	if err != nil {
		return fmt.Errorf("erro ao carregar template: %v", err)
	}

	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, templateData); err != nil {
		return fmt.Errorf("erro ao executar template: %v", err)
	}

	// Enviar email com o HTML
	mime := "MIME-Version: 1.0\n" +
		"Content-Type: text/html; charset=UTF-8\n\n" +
		emailBody.String()

	msg := []byte(fmt.Sprintf("Subject: %s\n%s", subject, mime))
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	return smtp.SendMail(addr, auth, config.From, []string{to}, msg)
}
