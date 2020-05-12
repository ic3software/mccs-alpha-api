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
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/l"

	"go.uber.org/zap"
)

var TransferHandler = newTransferHandler()

type transferHandler struct {
	once *sync.Once
}

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
		req, errs := handler.newTransferReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(req.InitiatorEntity.ID.Hex(), r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		err := logic.Transfer.CheckBalance(req.FromAccountNumber, req.ToAccountNumber, req.Amount)
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

		go logic.UserAction.ProposeTransfer(r.Header.Get("userID"), req)
		go email.Transfer.Initiate(req)
	}
}

func (handler *transferHandler) newTransferReq(r *http.Request) (*types.TransferReq, []error) {
	var body types.TransferUserReq
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
	return types.NewTransferReq(&body, initiatorEntity, receiverEntity)
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

func (handler *transferHandler) newSearchTransferQuery(r *http.Request) (*types.SearchTransferReq, []error) {
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
	var generateRespond = func(req *types.UpdateTransferReq, updated *types.Journal) *types.TransferRespond {
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
		req, errs := handler.newUpdateTransferReq(r)
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
			go logic.UserAction.AcceptTransfer(r.Header.Get("userID"), updated)
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

func (handler *transferHandler) newUpdateTransferReq(r *http.Request) (*types.UpdateTransferReq, []error) {
	transferID := mux.Vars(r)["transferID"]
	journal, err := logic.Transfer.FindByID(transferID)
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
	return types.NewUpdateTransferReq(r, journal, initiateEntity, fromEntity, toEntity)
}

func (handler *transferHandler) checkPermissions(req *types.UpdateTransferReq) error {
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

func (handler *transferHandler) checkBalances(req *types.UpdateTransferReq) error {
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
		reason := "The sender will exceed its credit limit so this transfer has been cancelled."
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
		reason := "The recipient will exceed its maximum positive balance threshold so this transfer has been cancelled."
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
	go email.Transfer.Accept(j)
	return updated, nil
}

func (handler *transferHandler) rejectTransfer(j *types.Journal, reason string) (*types.Journal, error) {
	updated, err := logic.Transfer.Cancel(j.TransferID, reason)
	if err != nil {
		return nil, err
	}
	go email.Transfer.Reject(j, reason)
	return updated, nil
}

func (handler *transferHandler) cancelTransfer(j *types.Journal, reason string) (*types.Journal, error) {
	updated, err := logic.Transfer.Cancel(j.TransferID, reason)
	if err != nil {
		return nil, err
	}
	go email.Transfer.Cancel(j, reason)
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
		Data *types.AdminTransferRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := handler.newAdminTransferReq(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		err := logic.Transfer.CheckBalance(req.PayerEntity.AccountNumber, req.PayeeEntity.AccountNumber, req.Amount)
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

		go logic.UserAction.AdminTransfer(r.Header.Get("userID"), journal)

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewJournalToAdminTransferRespond(journal)})
	}
}

func (handler *transferHandler) newAdminTransferReq(r *http.Request) (*types.AdminTransferReq, []error) {
	var body types.AdminTransferUserReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		if err == io.EOF {
			return nil, []error{errors.New("Please provide valid inputs.")}
		}
		return nil, []error{err}
	}
	payerEntity, err := logic.Entity.FindByAccountNumber(body.Payer)
	if err != nil {
		return nil, []error{err}
	}
	payeeEntity, err := logic.Entity.FindByAccountNumber(body.Payee)
	if err != nil {
		return nil, []error{err}
	}
	return types.NewAdminTransferReq(&body, payerEntity, payeeEntity)
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
