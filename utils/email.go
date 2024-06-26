package utils

import (
	"bytes"
	"crypto/tls"
	"github.com/Arxtect/Einstein/apps/archive/models"
	"github.com/Arxtect/Einstein/config"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/k3a/html2text"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	URL              string
	VerificationCode string
	FirstName        string
	Subject          string
	Amount           int64
	Balance          int64
}

// ? Email template parser

func ParseTemplateDir(dir string) (*template.Template, error) {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return template.ParseFiles(paths...)
}

func SendEmail(user *models.User, data *EmailData, emailTemp string) {
	configCopy := config.Env

	// Sender data.
	from := configCopy.EmailFrom
	smtpPass := configCopy.SMTPPass
	smtpUser := configCopy.SMTPUser
	to := user.Email
	smtpHost := configCopy.SMTPHost
	smtpPort := configCopy.SMTPPort

	var body bytes.Buffer

	tmpl, err := ParseTemplateDir("common/templates")
	if err != nil {
		log.Fatal("Could not parse template", err)
	}

	err = tmpl.ExecuteTemplate(&body, emailTemp, &data)
	if err != nil {
		log.Fatal("Could not execute template verificationCode", err)
	}

	m := gomail.NewMessage()

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", body.String())
	m.AddAlternative("text/plain", html2text.HTML2Text(body.String()))

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send Email
	if err := d.DialAndSend(m); err != nil {
		log.Fatal("Could not send email: ", err)
	}

}
