package email

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type balance struct{}

var Balance = &balance{}

// Non-zero balance notification

type NonZeroBalanceEmail struct {
	From time.Time
	To   time.Time
}

func (_ *balance) SendNonZeroBalanceEmail(input *NonZeroBalanceEmail) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.non_zero_balance_notification"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(viper.GetString("email_from"), viper.GetString("sendgrid.sender_email")),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("fromTime", input.From.Format("2006-01-02 15:04:05"))
	p.SetDynamicTemplateData("toTime", input.To.Format("2006-01-02 15:04:05"))
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.sendNonZeroBalanceEmail failed", zap.Error(err))
	}
}
