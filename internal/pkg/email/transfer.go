package email

import (
	"fmt"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util/l"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type transfer struct{}

var Transfer = &transfer{}

func (tr *transfer) Initiate(req *types.TransferReq) {
	url := viper.GetString("url") + "/pending-transfers"

	var body string
	if req.TransferDirection == constant.TransferDirection.Out {
		body = req.InitiatorEntityName + " wants to send " + fmt.Sprintf("%.2f", req.Amount) + " Credits to you. <a href=" + url + ">Click here to review this pending transaction</a>."
	}
	if req.TransferDirection == constant.TransferDirection.In {
		body = req.InitiatorEntityName + " wants to receive " + fmt.Sprintf("%.2f", req.Amount) + " Credits from you. <a href=" + url + ">Click here to review this pending transaction</a>."
	}

	d := emailData{
		receiver:      req.ReceiverEntityName,
		receiverEmail: req.ReceiverEmail,
		subject:       "OCN Transaction Requiring Your Approval",
		text:          body,
		html:          body,
	}
	err := e.send(d)
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

func (tr *transfer) Accept(info *TransferEmailInfo) {
	var body string
	if info.TransferDirection == "out" {
		body = info.ReceiverEntityName + " has accepted the transaction you initiated for -" + fmt.Sprintf("%.2f", info.Amount) + " Credits."
	} else {
		body = info.ReceiverEntityName + " has accepted the transaction you initiated for +" + fmt.Sprintf("%.2f", info.Amount) + " Credits."
	}
	d := emailData{
		receiver:      info.InitiatorEntityName,
		receiverEmail: info.InitiatorEmail,
		subject:       "OCN Transaction Accepted",
		text:          body,
		html:          body,
	}
	err := e.send(d)
	if err != nil {
		l.Logger.Error("email.Transfer.Accept failed", zap.Error(err))
	}
}

func (tr *transfer) Reject(info *TransferEmailInfo) {
	var body string
	if info.TransferDirection == "out" {
		body = info.ReceiverEntityName + " has rejected the transaction you initiated for -" + fmt.Sprintf("%.2f", info.Amount) + " Credits."
	} else {
		body = info.ReceiverEntityName + " has rejected the transaction you initiated for +" + fmt.Sprintf("%.2f", info.Amount) + " Credits."
	}

	if info.Reason != "" {
		body += "<br/><br/> Reason: <br/><br/>" + info.Reason
	}

	d := emailData{
		receiver:      info.InitiatorEntityName,
		receiverEmail: info.InitiatorEmail,
		subject:       "OCN Transaction Rejected",
		text:          body,
		html:          body,
	}
	err := e.send(d)
	if err != nil {
		l.Logger.Error("email.Transfer.Reject failed", zap.Error(err))
	}
}

func (tr *transfer) Cancel(info *TransferEmailInfo) {
	var body string
	if info.TransferDirection == "out" {
		body = info.InitiatorEntityName + " has cancelled the transaction it initiated for +" + fmt.Sprintf("%.2f", info.Amount) + " Credits."
	} else {
		body = info.InitiatorEntityName + " has cancelled the transaction it initiated for -" + fmt.Sprintf("%.2f", info.Amount) + " Credits."
	}

	if info.Reason != "" {
		body += "<br/><br/> Reason: <br/><br/>" + info.Reason
	}

	d := emailData{
		receiver:      info.ReceiverEntityName,
		receiverEmail: info.ReceiverEmail,
		subject:       "OCN Transaction Cancelled",
		text:          body,
		html:          body,
	}
	err := e.send(d)
	if err != nil {
		l.Logger.Error("email.Transfer.Cancel failed", zap.Error(err))
	}
}

func (tr *transfer) CancelBySystem(info *TransferEmailInfo) {
	body := "The system has cancelled the transaction you initiated with " + info.ReceiverEntityName + " for the following reason: " + info.Reason
	d := emailData{
		receiver:      info.InitiatorEntityName,
		receiverEmail: info.InitiatorEmail,
		subject:       "OCN Transaction Cancelled",
		text:          body,
		html:          body,
	}
	err := e.send(d)
	if err != nil {
		l.Logger.Error("email.Transfer.CancelBySystem failed", zap.Error(err))
	}
}
