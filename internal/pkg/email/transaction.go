package email

import (
	"fmt"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/spf13/viper"
)

type transaction struct{}

var Transaction = &transaction{}

func (tr *transaction) Initiate(proposal *types.TransferProposal) error {
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

// TO BE REMOVED

// type emailInfo struct {
// 	InitiatorEmail,
// 	InitiatorEntityName,
// 	ReceiverEmail,
// 	ReceiverEntityName string
// }

// func (tr *transaction) getEmailInfo(t *types.Transfer) *emailInfo {
// 	var initiatorEmail, initiatorEntityName, receiverEmail, receiverEntityName string
// 	if t.InitiatedBy == t.FromID {
// 		initiatorEntityName = t.FromEntityName
// 		initiatorEmail = t.FromEmail
// 		receiverEntityName = t.ToEntityName
// 		receiverEmail = t.ToEmail
// 	} else {
// 		initiatorEntityName = t.ToEntityName
// 		initiatorEmail = t.ToEmail
// 		receiverEntityName = t.FromEntityName
// 		receiverEmail = t.FromEmail
// 	}
// 	return &emailInfo{
// 		initiatorEmail,
// 		initiatorEntityName,
// 		receiverEmail,
// 		receiverEntityName,
// 	}
// }

// func (tr *transaction) Accept(t *types.Transfer) error {
// 	info := tr.getEmailInfo(t)

// 	var body string
// 	if t.InitiatedBy == t.FromID {
// 		body = info.ReceiverEntityName + " has accepted the transaction you initiated for -" + fmt.Sprintf("%.2f", t.Amount) + " Credits."
// 	} else {
// 		body = info.ReceiverEntityName + " has accepted the transaction you initiated for +" + fmt.Sprintf("%.2f", t.Amount) + " Credits."
// 	}

// 	d := emailData{
// 		receiver:      info.InitiatorEntityName,
// 		receiverEmail: info.InitiatorEmail,
// 		subject:       "OCN Transaction Accepted",
// 		text:          body,
// 		html:          body,
// 	}
// 	err := e.send(d)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (tr *transaction) Cancel(t *types.Transfer, reason string) error {
// 	info := tr.getEmailInfo(t)

// 	var body string
// 	if t.InitiatedBy == t.FromID {
// 		body = info.InitiatorEntityName + " has cancelled the transaction it initiated for +" + fmt.Sprintf("%.2f", t.Amount) + " Credits."
// 	} else {
// 		body = info.InitiatorEntityName + " has cancelled the transaction it initiated for -" + fmt.Sprintf("%.2f", t.Amount) + " Credits."
// 	}

// 	if reason != "" {
// 		body += "<br/><br/> Reason: <br/><br/>" + reason
// 	}

// 	d := emailData{
// 		receiver:      info.ReceiverEntityName,
// 		receiverEmail: info.ReceiverEmail,
// 		subject:       "OCN Transaction Cancelled",
// 		text:          body,
// 		html:          body,
// 	}
// 	err := e.send(d)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (tr *transaction) CancelBySystem(t *types.Transfer, reason string) error {
// 	info := tr.getEmailInfo(t)
// 	body := "The system has cancelled the transaction you initiated with " + info.ReceiverEntityName + " for the following reason: " + reason
// 	d := emailData{
// 		receiver:      info.InitiatorEntityName,
// 		receiverEmail: info.InitiatorEmail,
// 		subject:       "OCN Transaction Cancelled",
// 		text:          body,
// 		html:          body,
// 	}
// 	err := e.send(d)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (tr *transaction) Reject(t *types.Transfer) error {
// 	info := tr.getEmailInfo(t)

// 	var body string
// 	if t.InitiatedBy == t.FromID {
// 		body = info.ReceiverEntityName + " has rejected the transaction you initiated for -" + fmt.Sprintf("%.2f", t.Amount) + " Credits."
// 	} else {
// 		body = info.ReceiverEntityName + " has rejected the transaction you initiated for +" + fmt.Sprintf("%.2f", t.Amount) + " Credits."
// 	}

// 	d := emailData{
// 		receiver:      info.InitiatorEntityName,
// 		receiverEmail: info.InitiatorEmail,
// 		subject:       "OCN Transaction Rejected",
// 		text:          body,
// 		html:          body,
// 	}
// 	err := e.send(d)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
