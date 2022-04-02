package services

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type MailService struct {
	logger *zap.Logger
}

func NewMailService(logger *zap.Logger) *MailService {
	return &MailService{logger: logger}
}

func (ms *MailService) SendMail(userName string, userEmail string, code string) error {
	from := mail.NewEmail("DFS Team", "dfs.pk.proj@gmail.com")
	subject := "Email verification for DFS"
	to := mail.NewEmail(userName, userEmail)

	plainTextContent := fmt.Sprintf("Go there and verify your account: http://localhost/api/verify/%s", code)
	htmlContent := fmt.Sprintf("Go there and verify your account: <a href=\"http://localhost/api/verify/%s\">DFS Account verification</a>", code)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		return err
	} else {
		ms.logger.Debug("Response status code: {Code}", zap.Int("Code", response.StatusCode))
		ms.logger.Debug("Response body: {Body}", zap.String("Body", response.Body))
		ms.logger.Debug("Response headers: {Headers}", zap.Any("Headers", response.Headers))
		return nil
	}
}
