package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/api"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/ic3network/mccs-alpha-api/util"

	"go.uber.org/zap"
)

type transferHandler struct {
	once *sync.Once
}

var TransferHandler = newTransferHandler()

func newTransferHandler() *transferHandler {
	return &transferHandler{
		once: new(sync.Once),
	}
}

func (handler *transferHandler) RegisterRoutes(
	public *mux.Router,
	private *mux.Router,
	adminPublic *mux.Router,
	adminPrivate *mux.Router,
) {
	handler.once.Do(func() {
		private.Path("/transfers").HandlerFunc(handler.proposeTransfer()).Methods("POST")
		private.Path("/transfers").HandlerFunc(handler.searchTransfer()).Methods("GET")
		private.Path("/transfers/{transferID}").HandlerFunc(handler.updateTransfer()).Methods("PATCH")

		adminPrivate.Path("/transfers").HandlerFunc(handler.adminCreateTransfer()).Methods("POST")
		adminPrivate.Path("/transfers").HandlerFunc(handler.adminSearchTransfer()).Methods("GET")
		adminPrivate.Path("/transfers/{transferID}").HandlerFunc(handler.adminGetTransfer()).Methods("GET")
	})
}

// POST /transfers

func (handler *transferHandler) proposeTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.ProposeTransferRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newTransferReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(req.InitiatorEntity.ID.Hex(), r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		err := logic.Transfer.CheckBalance(req)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		journal, err := logic.Transfer.Propose(req)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.proposeTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewProposeTransferRespond(journal)})

		go func() {
			err := email.Transfer.Initiate(req)
			if err != nil {
				l.Logger.Error("email.Transfer.Initiate failed", zap.Error(err))
			}
		}()
	}
}

func (handler *transferHandler) newTransferReqBody(r *http.Request) (*types.TransferReqBody, []error) {
	var body types.TransferUserReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		if err == io.EOF {
			return nil, []error{errors.New("Please provide valid inputs.")}
		}
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
	return types.NewTransferReqBody(&body, initiatorEntity, receiverEntity)
}

// GET /transfers

func (handler *transferHandler) searchTransfer() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.TransferRespond `json:"data"`
		Meta meta                     `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newSearchTransferQuery(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(req.QueryingEntityID, r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		found, err := logic.Transfer.Search(req)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.searchTransfers failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: found.Transfers,
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

func (handler *transferHandler) newSearchTransferQuery(r *http.Request) (*types.SearchTransferReqBody, []error) {
	entity, err := logic.Entity.FindByStringID(r.URL.Query().Get("querying_entity_id"))
	if err != nil {
		return nil, []error{err}
	}
	req, errs := types.NewSearchTransferQuery(r, entity)
	if len(errs) > 0 {
		return nil, errs
	}
	return req, nil
}

// PATCH /transfers/{transferID}

func (handler *transferHandler) updateTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.TransferRespond `json:"data"`
	}
	var generateRespond = func(req *types.UpdateTransferReqBody, updated *types.Journal) *types.TransferRespond {
		t := &types.TransferRespond{
			TransferID:  req.TransferID,
			Description: req.Journal.Description,
			Amount:      req.Journal.Amount,
			CreatedAt:   &req.Journal.CreatedAt,
			Status:      updated.Status,
		}

		if util.ContainID(req.InitiateEntity.Users, req.LoggedInUserID) {
			t.IsInitiator = true
		}
		if util.ContainID(req.FromEntity.Users, req.LoggedInUserID) {
			t.Transfer = "out"
			t.AccountNumber = req.Journal.ToAccountNumber
			t.EntityName = req.Journal.ToEntityName
		} else {
			t.Transfer = "in"
			t.AccountNumber = req.Journal.FromAccountNumber
			t.EntityName = req.Journal.FromEntityName
		}
		if updated.Status == constant.Transfer.Completed {
			t.CompletedAt = &updated.UpdatedAt
		}

		return t
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newUpdateTransferReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		err := handler.checkPermissions(req)
		if err != nil {
			api.Respond(w, r, http.StatusUnauthorized, err)
			return
		}
		err = handler.checkBalances(req)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		var updated *types.Journal
		if req.Action == "accept" {
			updated, err = handler.acceptTransfer(req.Journal)
			if err != nil {
				l.Logger.Error("[Error] TransferHandler.updateTransfer failed:", zap.Error(err))
				api.Respond(w, r, http.StatusInternalServerError, err)
				return
			}
		}
		if req.Action == "reject" {
			updated, err = handler.rejectTransfer(req.Journal, req.Reason)
			if err != nil {
				l.Logger.Error("[Error] TransferHandler.updateTransfer failed:", zap.Error(err))
				api.Respond(w, r, http.StatusInternalServerError, err)
				return
			}
		}
		if req.Action == "cancel" {
			updated, err = handler.cancelTransfer(req.Journal, req.Reason)
			if err != nil {
				l.Logger.Error("[Error] TransferHandler.updateTransfer failed:", zap.Error(err))
				api.Respond(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		api.Respond(w, r, http.StatusOK, respond{generateRespond(req, updated)})
	}
}

func (handler *transferHandler) newUpdateTransferReqBody(r *http.Request) (*types.UpdateTransferReqBody, []error) {
	transferID := mux.Vars(r)["transferID"]
	journal, err := logic.Transfer.FindJournal(transferID)
	if err != nil {
		return nil, []error{err}
	}
	initiateEntity, err := logic.Entity.FindByAccountNumber(journal.InitiatedBy)
	if err != nil {
		return nil, []error{err}
	}
	fromEntity, err := logic.Entity.FindByAccountNumber(journal.FromAccountNumber)
	if err != nil {
		return nil, []error{err}
	}
	toEntity, err := logic.Entity.FindByAccountNumber(journal.ToAccountNumber)
	if err != nil {
		return nil, []error{err}
	}
	return types.NewUpdateTransferReqBody(r, journal, initiateEntity, fromEntity, toEntity)
}

func (handler *transferHandler) checkPermissions(req *types.UpdateTransferReqBody) error {
	if !util.ContainID(req.FromEntity.Users, req.LoggedInUserID) && !util.ContainID(req.ToEntity.Users, req.LoggedInUserID) {
		return errors.New("You don't have permission to perform this action.")
	}

	// If the logged in user is the owner of the initiate entity, then the user can only "cancel" the transfer.
	if util.ContainID(req.InitiateEntity.Users, req.LoggedInUserID) {
		if req.Action != "cancel" {
			return errors.New("You don't have permission to perform this action.")
		}
	} else {
		if req.Action != "accept" && req.Action != "reject" {
			return errors.New("You don't have permission to perform this action.")
		}
	}

	return nil
}

func (handler *transferHandler) checkBalances(req *types.UpdateTransferReqBody) error {
	fromAccount, err := logic.Account.FindByAccountNumber(req.Journal.FromAccountNumber)
	if err != nil {
		return err
	}
	toAccount, err := logic.Account.FindByAccountNumber(req.Journal.ToAccountNumber)
	if err != nil {
		return err
	}

	exceed, err := logic.BalanceLimit.IsExceedLimit(fromAccount.AccountNumber, fromAccount.Balance-req.Journal.Amount)
	if err != nil {
		return err
	}
	if exceed {
		reason := "The sender will exceed its credit limit so this tansfer has been cancelled."
		_, err = logic.Transfer.Cancel(req.Journal.TransferID, reason)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.updateTransfer failed:", zap.Error(err))
			return err
		}
		go func() {
			err := email.Transfer.CancelBySystem(req.Journal, reason)
			if err != nil {
				l.Logger.Error("[Error] email.Transfer.CancelBySystem failed:", zap.Error(err))
			}
		}()
		return errors.New(reason)
	}

	exceed, err = logic.BalanceLimit.IsExceedLimit(toAccount.AccountNumber, toAccount.Balance+req.Journal.Amount)
	if err != nil {
		return err
	}
	if exceed {
		reason := "The recipient will exceed its maximum positive balance threshold so this tansfer has been cancelled."
		_, err = logic.Transfer.Cancel(req.Journal.TransferID, reason)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.updateTransfer failed:", zap.Error(err))
			return err
		}
		go func() {
			err := email.Transfer.CancelBySystem(req.Journal, reason)
			if err != nil {
				l.Logger.Error("[Error] email.Transfer.CancelBySystem failed:", zap.Error(err))
			}
		}()
		return errors.New(reason)
	}

	return nil
}

func (handler *transferHandler) acceptTransfer(j *types.Journal) (*types.Journal, error) {
	updated, err := logic.Transfer.Accept(j)
	if err != nil {
		return nil, err
	}
	go func() {
		err := email.Transfer.Accept(j)
		if err != nil {
			l.Logger.Error("email.Transfer.Accept failed", zap.Error(err))
		}
	}()

	return updated, nil
}

func (handler *transferHandler) rejectTransfer(j *types.Journal, reason string) (*types.Journal, error) {
	updated, err := logic.Transfer.Cancel(j.TransferID, reason)
	if err != nil {
		return nil, err
	}
	go func() {
		err := email.Transfer.Reject(j, reason)
		if err != nil {
			l.Logger.Error("email.Transfer.Reject failed", zap.Error(err))
		}
	}()

	return updated, nil
}

func (handler *transferHandler) cancelTransfer(j *types.Journal, reason string) (*types.Journal, error) {
	updated, err := logic.Transfer.Cancel(j.TransferID, reason)
	if err != nil {
		return nil, err
	}
	go func() {
		err := email.Transfer.Cancel(j, reason)
		if err != nil {
			l.Logger.Error("email.Transfer.Cancel failed", zap.Error(err))
		}
	}()

	return updated, nil
}

// GET /admin/transfers

func (handler *transferHandler) adminSearchTransfer() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.AdminTransferRespond `json:"data"`
		Meta meta                          `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminSearchTransferQuery(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		found, err := logic.Transfer.AdminSearch(req)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.adminSearchTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{
			Data: found.Transfers,
			Meta: meta{
				TotalPages:      found.TotalPages,
				NumberOfResults: found.NumberOfResults,
			},
		})
	}
}

// POST /admin/transfers

func (handler *transferHandler) adminCreateTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.ProposeTransferRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newAdminTransferReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		err := logic.Transfer.CheckBalance(req)
		if err != nil {
			api.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		journal, err := logic.Transfer.Create(req)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.adminCreateTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewProposeTransferRespond(journal)})
	}
}

func (handler *transferHandler) newAdminTransferReqBody(r *http.Request) (*types.TransferReqBody, []error) {
	var body types.TransferUserReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		if err == io.EOF {
			return nil, []error{errors.New("Please provide valid inputs.")}
		}
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
	return types.NewTransferReqBody(&body, initiatorEntity, receiverEntity)
}

// GET /admin/transfers/{transferID}

func (handler *transferHandler) adminGetTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.AdminTransferRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewAdminGetTransfer(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		journal, err := logic.Transfer.AdminGetTransfer(req.TransferID)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.adminGetTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewJournalToAdminTransferRespond(journal)})
	}
}
