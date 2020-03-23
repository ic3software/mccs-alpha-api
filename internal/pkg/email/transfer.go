package email

import (
	"fmt"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/spf13/viper"
)

type transfer struct{}

var Transfer = &transfer{}

func (tr *transfer) Initiate(proposal *types.TransferProposal) error {
	url := viper.GetString("url") + "/pending_transactions"

	var body string
	if proposal.TransferType == constant.TransferType.Out {
		body = proposal.InitiatorEntityName + " wants to send " + fmt.Sprintf("%.2f", proposal.Amount) + " Credits to you. <a href=" + url + ">Click here to review this pending transaction</a>."
	}
	if proposal.TransferType == constant.TransferType.In {
		body = proposal.InitiatorEntityName + " wants to receive " + fmt.Sprintf("%.2f", proposal.Amount) + " Credits from you. <a href=" + url + ">Click here to review this pending transaction</a>."
	}

	d := emailData{
		receiver:      proposal.ReceiverEntityName,
		receiverEmail: proposal.ReceiverEmail,
		subject:       "OCN Transaction Requiring Your Approval",
		text:          body,
		html:          body,
	}
	err := e.send(d)
	if err != nil {
		return err
	}
	return nil
}

func (tr *transfer) Accept(j *types.Journal) error {
	info := tr.getEmailInfo(j)

	var body string
	if j.InitiatedBy == j.FromAccountNumber {
		body = info.ReceiverEntityName + " has accepted the transaction you initiated for -" + fmt.Sprintf("%.2f", j.Amount) + " Credits."
	} else {
		body = info.ReceiverEntityName + " has accepted the transaction you initiated for +" + fmt.Sprintf("%.2f", j.Amount) + " Credits."
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
		return err
	}
	return nil
}

func (tr *transfer) Reject(j *types.Journal) error {
	info := tr.getEmailInfo(j)

	var body string
	if j.InitiatedBy == j.FromAccountNumber {
		body = info.ReceiverEntityName + " has rejected the transaction you initiated for -" + fmt.Sprintf("%.2f", j.Amount) + " Credits."
	} else {
		body = info.ReceiverEntityName + " has rejected the transaction you initiated for +" + fmt.Sprintf("%.2f", j.Amount) + " Credits."
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
		return err
	}
	return nil
}

func (tr *transfer) Cancel(j *types.Journal, reason string) error {
	info := tr.getEmailInfo(j)

	var body string
	if j.InitiatedBy == j.FromAccountNumber {
		body = info.InitiatorEntityName + " has cancelled the transaction it initiated for +" + fmt.Sprintf("%.2f", j.Amount) + " Credits."
	} else {
		body = info.InitiatorEntityName + " has cancelled the transaction it initiated for -" + fmt.Sprintf("%.2f", j.Amount) + " Credits."
	}

	if reason != "" {
		body += "<br/><br/> Reason: <br/><br/>" + reason
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
		return err
	}
	return nil
}

func (tr *transfer) CancelBySystem(j *types.Journal, reason string) error {
	proposal := tr.getEmailInfo(j)

	body := "The system has cancelled the transaction you initiated with " + proposal.ReceiverEntityName + " for the following reason: " + reason
	d := emailData{
		receiver:      proposal.InitiatorEntityName,
		receiverEmail: proposal.InitiatorEmail,
		subject:       "OCN Transaction Cancelled",
		text:          body,
		html:          body,
	}
	err := e.send(d)
	if err != nil {
		return err
	}
	return nil
}

type emailInfo struct {
	InitiatorEmail,
	InitiatorEntityName,
	ReceiverEmail,
	ReceiverEntityName string
}

func (tr *transfer) getEmailInfo(j *types.Journal) *emailInfo {
	var initiatorEmail, initiatorEntityName, receiverEmail, receiverEntityName string
	if j.InitiatedBy == j.FromAccountNumber {
		initiatorEntityName = j.FromEntityName
		initiatorEmail = j.FromEmail
		receiverEntityName = j.ToEntityName
		receiverEmail = j.ToEmail
	} else {
		initiatorEntityName = j.ToEntityName
		initiatorEmail = j.ToEmail
		receiverEntityName = j.FromEntityName
		receiverEmail = j.FromEmail
	}
	return &emailInfo{
		initiatorEmail,
		initiatorEntityName,
		receiverEmail,
		receiverEntityName,
	}
}
