package types

import (
	"errors"

	"github.com/ic3network/mccs-alpha-api/global/constant"
)

type TransferProposal struct {
	TransferType string
	Amount       float64
	Description  string

	InitiatorAccountNumber string
	InitiatorEmail         string
	InitiatorEntityName    string

	ReceiverAccountNumber string
	ReceiverEmail         string
	ReceiverEntityName    string

	FromAccountNumber string
	FromEmail         string
	FromEntityName    string
	FromStatus        string

	ToAccountNumber string
	ToEmail         string
	ToEntityName    string
	ToStatus        string
}

func NewTransferProposal(req *TransferReqBody, initiatorEntity, receiverEntity *Entity) (*TransferProposal, []error) {
	proposal := &TransferProposal{
		InitiatorAccountNumber: initiatorEntity.AccountNumber,
		InitiatorEmail:         initiatorEntity.Email,
		InitiatorEntityName:    initiatorEntity.EntityName,
		ReceiverAccountNumber:  receiverEntity.AccountNumber,
		ReceiverEmail:          receiverEntity.Email,
		ReceiverEntityName:     receiverEntity.EntityName,
		Amount:                 req.Amount,
		Description:            req.Description,
	}

	if req.Transfer == constant.TransferType.Out {
		proposal.TransferType = constant.TransferType.Out

		proposal.FromAccountNumber = initiatorEntity.AccountNumber
		proposal.FromEmail = initiatorEntity.Email
		proposal.FromEntityName = initiatorEntity.EntityName
		proposal.FromStatus = initiatorEntity.Status

		proposal.ToAccountNumber = receiverEntity.AccountNumber
		proposal.ToEmail = receiverEntity.Email
		proposal.ToEntityName = receiverEntity.EntityName
		proposal.ToStatus = receiverEntity.Status
	}

	if req.Transfer == constant.TransferType.In {
		proposal.TransferType = constant.TransferType.In

		proposal.FromAccountNumber = receiverEntity.AccountNumber
		proposal.FromEmail = receiverEntity.Email
		proposal.FromEntityName = receiverEntity.EntityName
		proposal.FromStatus = receiverEntity.Status

		proposal.ToAccountNumber = initiatorEntity.AccountNumber
		proposal.ToEmail = initiatorEntity.Email
		proposal.ToEntityName = initiatorEntity.EntityName
		proposal.ToStatus = initiatorEntity.Status
	}

	return proposal, proposal.validate()
}

func (proposal *TransferProposal) validate() []error {
	errs := []error{}

	// Only allow transfers with accounts that also have "trading-accepted" status
	if proposal.FromStatus != constant.Trading.Accepted {
		errs = append(errs, errors.New("Sender is not a trading member. Transfers can only be made when both entities have trading member status."))
	} else if proposal.ToStatus != constant.Trading.Accepted {
		errs = append(errs, errors.New("Recipient is not a trading member. Transfers can only be made when both entities have trading member status."))
	}

	// Check if the user is doing the transaction to himself.
	if proposal.FromAccountNumber == proposal.ToAccountNumber {
		errs = append(errs, errors.New("You cannot create a transaction with yourself."))
	}

	return errs
}
