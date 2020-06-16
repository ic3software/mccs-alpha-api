package email

import (
	"fmt"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var e *Email

type Email struct{}

func (_ *Email) newEmail(templateID string) *mail.SGMailV3 {
	m := mail.NewV3Mail()
	e := mail.NewEmail(viper.GetString("email_from"), viper.GetString("sendgrid.sender_email"))
	m.SetFrom(e)
	m.SetTemplateID(templateID)
	return m
}

func (_ *Email) send(m *mail.SGMailV3) error {
	request := sendgrid.GetRequest(viper.GetString("sendgrid.key"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	_, err := sendgrid.API(request)
	return err
}

// Welcome message

type WelcomeEmail struct {
	EntityName string
	Email      string
	Receiver   string
}

func Welcome(input *WelcomeEmail) {
	e.welcome(input)
}
func (_ *Email) welcome(input *WelcomeEmail) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.welcome"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(input.Receiver+" ", input.Email),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("entityName", input.EntityName)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Welcome failed", zap.Error(err))
	}
}

// Signup notification

type SignupNotificationEmail struct {
	EntityName   string
	ContactEmail string
}

func Signup(input *SignupNotificationEmail) {
	e.signup(input)
}
func (_ *Email) signup(input *SignupNotificationEmail) {
	if !viper.GetBool("receive_email.signup_notifications") {
		return
	}

	m := e.newEmail(viper.GetString("sendgrid.template_id.signup_notification"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(viper.GetString("email_from"), viper.GetString("sendgrid.sender_email")),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("entityName", input.EntityName)
	p.SetDynamicTemplateData("contactEmail", input.ContactEmail)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Signup failed", zap.Error(err))
	}
}

// Password reset

type PasswordResetEmail struct {
	Receiver      string
	ReceiverEmail string
	Token         string
}

func PasswordReset(input *PasswordResetEmail) {
	e.passwordReset(input)
}
func (_ *Email) passwordReset(input *PasswordResetEmail) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.password_reset"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(input.Receiver+" ", input.ReceiverEmail),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("serverAddress", viper.GetString("url"))
	p.SetDynamicTemplateData("token", input.Token)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.PasswordReset failed", zap.Error(err))
	}
}

// Admin password reset

type AdminPasswordResetEmail struct {
	Receiver      string
	ReceiverEmail string
	Token         string
}

func AdminPasswordReset(input *AdminPasswordResetEmail) {
	e.adminPasswordReset(input)
}
func (_ *Email) adminPasswordReset(input *AdminPasswordResetEmail) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.admin_reset_password"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(input.Receiver+" ", input.ReceiverEmail),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("serverAddress", viper.GetString("url"))
	p.SetDynamicTemplateData("token", input.Token)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.AdminPasswordReset failed", zap.Error(err))
	}
}

// Trade contact

type TradeContactEmail struct {
	Receiver      string
	ReceiverEmail string
	ReplyToName   string
	ReplyToEmail  string
	Body          string
}

func TradeContact(input *TradeContactEmail) {
	e.tradeContact(input)
}
func (_ *Email) tradeContact(input *TradeContactEmail) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.trade_contact"))
	replyToEmail := mail.NewEmail(input.ReplyToName, input.ReplyToEmail)
	m.SetReplyTo(replyToEmail)

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(input.Receiver+" ", input.ReceiverEmail),
	}
	if viper.GetBool("receive_email.trade_contact_emails") {
		tos = append(tos, mail.NewEmail(viper.GetString("email_from"), viper.GetString("sendgrid.sender_email")))
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("body", input.Body)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.TradeContact failed", zap.Error(err))
	}
}

// DailyEmailList

type DailyMatchNotification struct {
	Entity      *types.Entity
	MatchedTags *types.MatchedTags
}

func DailyMatch(input *DailyMatchNotification) {
	e.dailyMatch(input)
}
func (_ *Email) dailyMatch(input *DailyMatchNotification) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.daily_match_notification"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(input.Entity.Name+" ", input.Entity.Email),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("matchedOffers", input.MatchedTags.MatchedOffers)
	p.SetDynamicTemplateData("matchedWants", input.MatchedTags.MatchedWants)
	p.SetDynamicTemplateData("lastNotificationSentDate", fmt.Sprintf("%d", input.Entity.LastNotificationSentDate.UTC().Unix()))
	p.SetDynamicTemplateData("url", viper.GetString("url"))
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.DailyMatch failed", zap.Error(err))
	}
}
