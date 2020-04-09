package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

func NewTransferReqBody(r *http.Request) (*types.TransferReqBody, []error) {
	var body struct {
		Transfer               string  `json:"transfer"`
		InitiatorAccountNumber string  `json:"initiator"`
		ReceiverAccountNumber  string  `json:"receiver"`
		Amount                 float64 `json:"amount"`
		Description            string  `json:"description"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		return nil, []error{err}
	}

	initiatorEntity, err := logic.Entity.FindByAccountNumber(body.InitiatorAccountNumber)
	if err != nil {
		return nil, []error{err}
	}
	receiverEntity, err := logic.Entity.FindByAccountNumber(body.ReceiverAccountNumber)
	if err != nil {
		return nil, []error{err}
	}

	req := &types.TransferReqBody{
		TransferType:           body.Transfer,
		Amount:                 body.Amount,
		Description:            body.Description,
		InitiatorAccountNumber: initiatorEntity.AccountNumber,
		InitiatorEmail:         initiatorEntity.Email,
		InitiatorEntityName:    initiatorEntity.EntityName,
		ReceiverAccountNumber:  receiverEntity.AccountNumber,
		ReceiverEmail:          receiverEntity.Email,
		ReceiverEntityName:     receiverEntity.EntityName,
		InitiatorEntity:        initiatorEntity,
		ReceiverEntity:         receiverEntity,
	}

	if req.TransferType == constant.TransferType.Out {
		req.FromAccountNumber = initiatorEntity.AccountNumber
		req.FromEmail = initiatorEntity.Email
		req.FromEntityName = initiatorEntity.EntityName
		req.FromStatus = initiatorEntity.Status

		req.ToAccountNumber = receiverEntity.AccountNumber
		req.ToEmail = receiverEntity.Email
		req.ToEntityName = receiverEntity.EntityName
		req.ToStatus = receiverEntity.Status
	}

	if req.TransferType == constant.TransferType.In {
		req.FromAccountNumber = receiverEntity.AccountNumber
		req.FromEmail = receiverEntity.Email
		req.FromEntityName = receiverEntity.EntityName
		req.FromStatus = receiverEntity.Status

		req.ToAccountNumber = initiatorEntity.AccountNumber
		req.ToEmail = initiatorEntity.Email
		req.ToEntityName = initiatorEntity.EntityName
		req.ToStatus = initiatorEntity.Status
	}

	return req, req.Validate()
}

func NewUpdateTransferReqBody(r *http.Request) (*types.UpdateTransferReqBody, []error) {
	var body struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		if err == io.EOF {
			return nil, []error{errors.New("Please provide valid inputs.")}
		}
		return nil, []error{err}
	}

	var req types.UpdateTransferReqBody

	req.TransferID = mux.Vars(r)["transferID"]
	req.LoggedInUserID = r.Header.Get("userID")
	req.Action = body.Action
	req.Reason = body.Reason

	journal, err := logic.Transfer.FindJournal(req.TransferID)
	if err != nil {
		return nil, []error{err}
	}
	req.Journal = journal

	initiateEntity, err := logic.Entity.FindByAccountNumber(req.Journal.InitiatedBy)
	if err != nil {
		return nil, []error{err}
	}
	req.InitiateEntity = initiateEntity

	fromEntity, err := logic.Entity.FindByAccountNumber(req.Journal.FromAccountNumber)
	if err != nil {
		return nil, []error{err}
	}
	req.FromEntity = fromEntity

	toEntity, err := logic.Entity.FindByAccountNumber(req.Journal.ToAccountNumber)
	if err != nil {
		return nil, []error{err}
	}
	req.ToEntity = toEntity

	return &req, req.Validate()
}
