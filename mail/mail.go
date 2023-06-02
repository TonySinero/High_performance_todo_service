package mail

import (
	"fmt"
	"net/smtp"
	"newFeatures/models"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	Host    = "smtp.gmail.com"
	Port    = "587"
	Subject = "Todo Service"
)

func SendEmail(post *models.Post) {
	auth := smtp.PlainAuth("", os.Getenv("POST_FROM"), os.Getenv("POST_PASSWORD"), Host)

	from := os.Getenv("POST_FROM")
	to := post.Email
	smtpHost := Host
	smtpPort := Port

	msg := fmt.Sprintf("Dear client, your current password is: %s.", post.Password)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, Subject, msg)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		logrus.Errorf("Error while sending email to %s: %s", to, err)
		return
	}

	logrus.Infof("Email for %s sent successfully!", to)
}
