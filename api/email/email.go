package email

import (
	"bytes"
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

	// Log das configurações (remova senhas em produção)
	fmt.Printf("Configuração de Email:\nFrom: %s\nHost: %s\nPort: %s\n",
		config.From, config.Host, config.Port)

	auth := smtp.PlainAuth("", config.From, config.Password, config.Host)

	templateData := struct {
		VerificationLink string
	}{
		VerificationLink: fmt.Sprintf("http://localhost:8080/verify-email?token=%s", token),
	}

	// Log do template
	fmt.Printf("Tentando carregar template em: %s\n", "view/email_templates/verification.html")

	tmpl, err := template.ParseFiles("view/email_templates/verification.html")
	if err != nil {
		fmt.Printf("Erro ao carregar template: %v\n", err)
		return fmt.Errorf("erro ao carregar template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, templateData); err != nil {
		fmt.Printf("Erro ao executar template: %v\n", err)
		return fmt.Errorf("erro ao executar template: %v", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: Confirme seu email - Superviso\n"
	msg := []byte(subject + mime + body.String())

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	fmt.Printf("Tentando enviar email para %s via %s\n", to, addr)

	if err := smtp.SendMail(addr, auth, config.From, []string{to}, msg); err != nil {
		fmt.Printf("Erro ao enviar email: %v\n", err)
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	fmt.Printf("Email enviado com sucesso para %s\n", to)
	return nil
}
