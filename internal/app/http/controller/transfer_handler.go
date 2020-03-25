package controller

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
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
		private.Path("/transfers").HandlerFunc(handler.searchTransfers()).Methods("GET")
		private.Path("/transfers/{transferID}").HandlerFunc(handler.updateTransfer()).Methods("PATCH")
	})
}

func (handler *transferHandler) proposeTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.ProposeTransferRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := api.NewTransferReqBody(r)
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

		api.Respond(w, r, http.StatusOK, respond{Data: api.NewProposeTransferRespond(journal)})

		go func() {
			err := email.Transfer.Initiate(req)
			if err != nil {
				l.Logger.Error("email.Transfer.Initiate failed", zap.Error(err))
			}
		}()
	}
}

func (handler *transferHandler) searchTransfers() func(http.ResponseWriter, *http.Request) {
	type meta struct {
		NumberOfResults int `json:"numberOfResults"`
		TotalPages      int `json:"totalPages"`
	}
	type respond struct {
		Data []*types.Transfer `json:"data"`
		Meta meta              `json:"meta"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		query, errs := api.NewSearchTransferQuery(r.URL.Query())
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		if !UserHandler.IsEntityBelongsToUser(query.QueryingEntityID, r.Header.Get("userID")) {
			api.Respond(w, r, http.StatusForbidden, api.ErrPermissionDenied)
			return
		}

		found, err := logic.Transfer.Search(query)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.searchTransfers failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		if err != nil {
			l.Logger.Error("[Error] EntityHandler.searchEntity failed:", zap.Error(err))
			api.Respond(w, r, http.StatusBadRequest, err)
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

func (handler *transferHandler) updateTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.Transfer `json:"data"`
	}
	var generateRespond = func(req *types.UpdateTransferReqBody, updated *types.Journal) *types.Transfer {
		t := &types.Transfer{
			TransferID:  req.TransferID,
			Description: req.Journal.Description,
			Amount:      req.Journal.Amount,
			CreatedAt:   req.Journal.CreatedAt,
			Status:      updated.Status,
			CompletedAt: updated.CompletedAt,
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

		return t
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := api.NewUpdateTransferReqBody(r)
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

	exceed, err := logic.BalanceLimit.IsExceedLimit(fromAccount.ID, fromAccount.Balance-req.Journal.Amount)
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

	exceed, err = logic.BalanceLimit.IsExceedLimit(toAccount.ID, toAccount.Balance+req.Journal.Amount)
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
