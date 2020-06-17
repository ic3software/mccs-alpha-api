package email

import (
	"fmt"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type transfer struct{}

var Transfer = &transfer{}

// Transfer initiated

func (tr *transfer) Initiate(req *types.TransferReq) {
	url := viper.GetString("url") + "/pending-transfers"

	var action string
	if req.TransferDirection == constant.TransferDirection.Out {
		action = "send " + fmt.Sprintf("%.2f", req.Amount) + " Credits to you"
	}
	if req.TransferDirection == constant.TransferDirection.In {
		action = "receive " + fmt.Sprintf("%.2f", req.Amount) + " Credits from you"
	}

	m := e.newEmail(viper.GetString("sendgrid.template_id.transfer_initiated"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(req.ReceiverEntityName+" ", req.ReceiverEmail),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("initiatorEntityName", req.InitiatorEntityName)
	p.SetDynamicTemplateData("action", action)
	p.SetDynamicTemplateData("url", url)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Transfer.Initiate failed", zap.Error(err))
	}
}

type TransferEmailInfo struct {
	TransferDirection   string
	InitiatorEmail      string
	InitiatorEntityName string
	ReceiverEmail       string
	ReceiverEntityName  string
	Reason              string
	Amount              float64
}

// Transfer accepted

func (tr *transfer) Accept(info *TransferEmailInfo) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.transfer_accepted"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(info.InitiatorEntityName+" ", info.InitiatorEmail),
	}
	p.AddTos(tos...)

	if info.TransferDirection == "out" {
		p.SetDynamicTemplateData("transferDirection", "-")
	} else {
		p.SetDynamicTemplateData("transferDirection", "+")
	}
	p.SetDynamicTemplateData("receiverEntityName", info.ReceiverEntityName)
	p.SetDynamicTemplateData("amount", fmt.Sprintf("%.2f", info.Amount))
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Transfer.Accept failed", zap.Error(err))
	}
}

// Transfer rejected

func (tr *transfer) Reject(info *TransferEmailInfo) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.transfer_rejected"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(info.InitiatorEntityName+" ", info.InitiatorEmail),
	}
	p.AddTos(tos...)

	if info.TransferDirection == "out" {
		p.SetDynamicTemplateData("transferDirection", "-")
	} else {
		p.SetDynamicTemplateData("transferDirection", "+")
	}
	p.SetDynamicTemplateData("receiverEntityName", info.ReceiverEntityName)
	p.SetDynamicTemplateData("amount", fmt.Sprintf("%.2f", info.Amount))
	p.SetDynamicTemplateData("reason", info.Reason)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Transfer.Reject failed", zap.Error(err))
	}
}

// Transfer cancelled

func (tr *transfer) Cancel(info *TransferEmailInfo) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.transfer_cancelled"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(info.ReceiverEntityName+" ", info.ReceiverEmail),
	}
	p.AddTos(tos...)

	if info.TransferDirection == "out" {
		p.SetDynamicTemplateData("transferDirection", "+")
	} else {
		p.SetDynamicTemplateData("transferDirection", "-")
	}
	p.SetDynamicTemplateData("initiatorEntityName", info.InitiatorEntityName)
	p.SetDynamicTemplateData("amount", fmt.Sprintf("%.2f", info.Amount))
	p.SetDynamicTemplateData("reason", info.Reason)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Transfer.Cancel failed", zap.Error(err))
	}
}

// Transfer cancelled by system

func (tr *transfer) CancelBySystem(info *TransferEmailInfo) {
	m := e.newEmail(viper.GetString("sendgrid.template_id.transfer_cancelled_by_system"))

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(info.InitiatorEntityName+" ", info.InitiatorEmail),
	}
	p.AddTos(tos...)

	p.SetDynamicTemplateData("receiverEntityName", info.ReceiverEntityName)
	p.SetDynamicTemplateData("reason", info.Reason)
	m.AddPersonalizations(p)

	err := e.send(m)
	if err != nil {
		l.Logger.Error("email.Transfer.Cancel failed", zap.Error(err))
	}
}
