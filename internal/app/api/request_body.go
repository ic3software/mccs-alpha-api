package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewSignupReqBody(r *http.Request) (*types.SignupReqBody, error) {
	var req types.SignupReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func NewLoginReqBody(r *http.Request) (*types.LoginReqBody, error) {
	var req types.LoginReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func NewUpdateUserEntityReqBody(r *http.Request) (*types.UpdateUserEntityReqBody, error) {
	var req types.UpdateUserEntityReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	req.Offers, req.Wants = util.FormatTags(req.Offers), util.FormatTags(req.Wants)
	return &req, nil
}

func NewAddToFavoriteReqBody(r *http.Request) (*types.AddToFavoriteReqBody, error) {
	var req types.AddToFavoriteReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func NewEmailReqBody(r *http.Request) (*types.EmailReqBody, error) {
	var req types.EmailReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func NewTransferReqBody(r *http.Request) (*types.TransferReqBody, []error) {
	var req types.TransferReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.Validate()
}

func NewUpdateTransferReqBody(r *http.Request) (*types.UpdateTransferReqBody, []error) {
	var body struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		return nil, []error{err}
	}

	var req types.UpdateTransferReqBody

	req.Action = body.Action
	req.Reason = body.Reason
	req.TransferID = mux.Vars(r)["searchEntityID"]

	journal, err := logic.Transfer.FindJournal(req.TransferID)
	if err != nil {
		return nil, []error{err}
	}
	req.Journal = journal

	return &req, req.Validate()
}

// Admin

func NewAdminUpdateEntityReqBody(r *http.Request) (*types.AdminUpdateEntityReqBody, error) {
	var req types.AdminUpdateEntityReqBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}

	vars := mux.Vars(r)
	entityID, err := primitive.ObjectIDFromHex(vars["entityID"])
	if err != nil {
		return nil, err
	}
	req.EntityID = entityID
	req.Offers, req.Wants, req.Categories = util.FormatTags(req.Offers), util.FormatTags(req.Wants), util.FormatTags(req.Categories)

	return &req, nil
}
