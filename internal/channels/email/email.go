package email

import (
	"context"

	"github.com/mrz1836/postmark"
	"github.com/retail-core/notification-service/internal/config"
	"github.com/retail-core/notification-service/internal/dispatcher"
	"github.com/retail-core/notification-service/internal/logger"
	"go.uber.org/zap"
)

type EmailHandler struct {
	client           *postmark.Client
	fromEmail        string
	fromName         string
	defaultRecipient *string
}

func NewPostmarkHandler(cfg *config.Config) *EmailHandler {

	var defaultRecipient *string
	if cfg.Environment != "production" {
		defaultRecipient = &cfg.POSTMARKToEmail
	}

	return &EmailHandler{
		client:           postmark.NewClient(cfg.POSTMARKServerToken, cfg.POSTMARKAccountToken),
		fromEmail:        cfg.POSTMARKFromEmail,
		fromName:         cfg.POSTMARKFromEmail,
		defaultRecipient: defaultRecipient,
	}
}

func (p *EmailHandler) Send(notification dispatcher.Notification) error {
	log := logger.L()

	recipientEmail := p.resolveRecipient(notification)

	var templateModel map[string]any

	if notification.Template == "staff-invitation" {
		templateModel = map[string]any{
			"product_url":    "www.jarabiz.netlify.com",
			"product_name":   "",
			"username":        notification.Data["username"],
			"otp":             notification.Data["otp"],
			"company_name":    "Jara-Biz",
			"company_address": "Benin City, Nigeria",
			"store_name":      notification.Data["store_name"],
			"temporary_password":       notification.Data["temporary_password"],
		}
	} else {
		templateModel = map[string]any{
			"product_url":     "www.jarabiz.netlify.com",
			"product_name":    "",
			"username":        notification.Data["username"],
			"otp":             notification.Data["otp"],
			"company_name":    "Jara-Biz",
			"company_address": "Benin City, Nigeria",
		}
	}

	email := postmark.TemplatedEmail{
		From:          p.fromEmail,
		To:            recipientEmail,
		TemplateAlias: notification.Template,
		TemplateModel: templateModel,
	}

	res, err := p.client.SendTemplatedEmail(context.Background(), email)

	if err != nil {
		log.Error("Failed to send email via Postmark", zap.Error(err))
		return err
	}

	log.Info("Email sent via Postmark", zap.String("to", notification.To), zap.String("messageID", res.MessageID))
	return nil
}

func (p *EmailHandler) resolveRecipient(n dispatcher.Notification) string {

	return n.To

	// if p.defaultRecipient != nil && *p.defaultRecipient != "" {
	// 	return *p.defaultRecipient
	// }
	// if n.To != "" {
	// 	return n.To
	// }
	// return p.fromEmail
}
