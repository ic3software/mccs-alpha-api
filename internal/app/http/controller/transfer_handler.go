package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/logic"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/api"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/email"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/flash"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/jsonerror"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/l"
	"github.com/unrolled/render"

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

		private.Path("/transaction/cancelPropose").HandlerFunc(handler.cancelPropose()).Methods("GET")
		private.Path("/pending_transactions").HandlerFunc(handler.pendingTransactionsPage()).Methods("GET")
		private.Path("/api/accountBalance").HandlerFunc(handler.getBalance()).Methods("GET")
		private.Path("/api/pendingTransactions").HandlerFunc(handler.pendingTransactions()).Methods("GET")
		private.Path("/api/acceptTransaction").HandlerFunc(handler.acceptTransaction()).Methods("POST")
		private.Path("/api/cancelTransaction").HandlerFunc(handler.cancelTransaction()).Methods("POST")
		private.Path("/api/rejectTransaction").HandlerFunc(handler.rejectTransaction()).Methods("POST")
		private.Path("/api/recentTransactions").HandlerFunc(handler.recentTransactions()).Methods("GET")
	})
}

func (handler *transferHandler) proposeTransfer() func(http.ResponseWriter, *http.Request) {
	type respond struct {
		Data *types.ProposeTransferRespond `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, errs := types.NewTransferReqBody(r)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		initiatorEntity, err := logic.Entity.FindByAccountNumber(req.InitiatorAccountNumber)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.proposeTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		receiverEntity, err := logic.Entity.FindByAccountNumber(req.ReceiverAccountNumber)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.proposeTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		proposal, errs := types.NewTransferProposal(req, initiatorEntity, receiverEntity)
		if len(errs) > 0 {
			api.Respond(w, r, http.StatusBadRequest, errs)
			return
		}

		journal, err := logic.Transfer.Propose(proposal)
		if err != nil {
			l.Logger.Error("[Error] TransferHandler.proposeTransfer failed:", zap.Error(err))
			api.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		api.Respond(w, r, http.StatusOK, respond{Data: types.NewProposeTransferRespond(journal)})

		go func() {
			err := email.Transaction.Initiate(proposal)
			if err != nil {
				l.Logger.Error("email.Transaction.Initiate failed", zap.Error(err))
			}
		}()
	}
}

// TO BE REMOVED

type formData struct {
	Type        string // "send" or "receive"
	Email       string
	Amount      float64
	Description string
}
type response struct {
	FormData      formData
	CurBalance    float64
	MaxNegBalance float64
}

func (handler *transferHandler) cancelPropose() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		flash.Info(w, "No transfer has been initiated from your account.")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (handler *transferHandler) getBalance() func(http.ResponseWriter, *http.Request) {
	type response struct {
		Balance float64
	}
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := AccountHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("TransferHandler.getBalance failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res := response{Balance: account.Balance}
		js, err := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (handler *transferHandler) pendingTransactions() func(http.ResponseWriter, *http.Request) {
	type response struct {
		Transactions []*types.Transfer
	}
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := AccountHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("TransferHandler.pendingTransactions failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		transactions, err := logic.Transfer.FindPendings(account.ID)
		if err != nil {
			l.Logger.Error("TransferHandler.pendingTransactions failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res := response{Transactions: transactions}
		js, err := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (handler *transferHandler) recentTransactions() func(http.ResponseWriter, *http.Request) {
	type response struct {
		Transactions []*types.Transfer
	}
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := AccountHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("TransferHandler.recentTransactions failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		transactions, err := logic.Transfer.FindRecent(account.ID)
		if err != nil {
			l.Logger.Error("TransferHandler.recentTransactions failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res := response{Transactions: transactions}
		js, err := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (handler *transferHandler) isInitiatedStatus(w http.ResponseWriter, t *types.Transfer) (bool, error) {
	type response struct {
		Error string `json:"error"`
	}

	if t.Status == constant.Transfer.Completed {
		js, err := json.Marshal(response{Error: "The transaction has already been completed by the counterparty."})
		if err != nil {
			return false, err
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return false, nil
	} else if t.Status == constant.Transfer.Cancelled {
		js, err := json.Marshal(response{Error: "The transaction has already been cancelled by the counterparty."})
		if err != nil {
			return false, err
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return false, nil
	}

	return true, nil
}

func (handler *transferHandler) cancelTransaction() func(http.ResponseWriter, *http.Request) {
	type request struct {
		TransactionID uint   `json:"id"`
		Reason        string `json:"reason"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		account, err := AccountHandler.FindByUserID(r.Header.Get("userID"))
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		transaction, err := logic.Transfer.Find(req.TransactionID)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shouldContinue, err := handler.isInitiatedStatus(w, transaction)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !shouldContinue {
			return
		}

		if account.ID != transaction.InitiatedBy {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = logic.Transfer.Cancel(req.TransactionID, req.Reason)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

		go func() {
			err := email.Transaction.Cancel(transaction, req.Reason)
			if err != nil {
				l.Logger.Error("email.Transaction.Cancel failed", zap.Error(err))
			}
		}()
	}
}

func (handler *transferHandler) rejectTransaction() func(http.ResponseWriter, *http.Request) {
	type request struct {
		TransactionID uint   `json:"id"`
		Reason        string `json:"reason"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("TransferHandler.rejectTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		transaction, err := logic.Transfer.Find(req.TransactionID)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shouldContinue, err := handler.isInitiatedStatus(w, transaction)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !shouldContinue {
			return
		}

		err = logic.Transfer.Cancel(req.TransactionID, req.Reason)
		if err != nil {
			l.Logger.Error("TransferHandler.rejectTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

		go func() {
			err := email.Transaction.Reject(transaction)
			if err != nil {
				l.Logger.Error("email.Transaction.Reject failed", zap.Error(err))
			}
		}()
	}
}

func (handler *transferHandler) acceptTransaction() func(http.ResponseWriter, *http.Request) {
	type request struct {
		TransactionID uint `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		transaction, err := logic.Transfer.Find(req.TransactionID)
		if err != nil {
			l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shouldContinue, err := handler.isInitiatedStatus(w, transaction)
		if err != nil {
			l.Logger.Error("TransferHandler.cancelTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !shouldContinue {
			return
		}

		from, err := logic.Account.FindByID(transaction.FromID)
		if err != nil {
			l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		to, err := logic.Account.FindByID(transaction.ToID)
		if err != nil {
			l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check the account balance.
		exceed, err := logic.BalanceLimit.IsExceedLimit(from.ID, from.Balance-transaction.Amount)
		if err != nil {
			l.Logger.Info("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exceed {
			reason := "The sender will exceed its credit limit so this transaction has been cancelled."
			err = logic.Transfer.Cancel(req.TransactionID, reason)
			if err != nil {
				l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err := jsonerror.New("1", reason)
			render.New().JSON(w, http.StatusInternalServerError, err.Render())
			go func() {
				err := email.Transaction.CancelBySystem(transaction, reason)
				if err != nil {
					l.Logger.Error("email.Transaction.Cancel failed", zap.Error(err))
				}
			}()
			return
		}
		exceed, err = logic.BalanceLimit.IsExceedLimit(to.ID, to.Balance+transaction.Amount)
		if err != nil {
			l.Logger.Info("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exceed {
			reason := "The recipient will exceed its maximum positive balance threshold so this transaction has been cancelled."
			err = logic.Transfer.Cancel(req.TransactionID, reason)
			if err != nil {
				l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err := jsonerror.New("2", reason)
			render.New().JSON(w, http.StatusInternalServerError, err.Render())
			go func() {
				err := email.Transaction.CancelBySystem(transaction, reason)
				if err != nil {
					l.Logger.Error("email.Transaction.Cancel failed", zap.Error(err))
				}
			}()
			return
		}

		err = logic.Transfer.Accept(transaction.ID, from.ID, to.ID, transaction.Amount)
		if err != nil {
			l.Logger.Error("TransferHandler.acceptTransaction failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

		go func() {
			err := email.Transaction.Accept(transaction)
			if err != nil {
				l.Logger.Error("email.Transaction.Accept failed", zap.Error(err))
			}
		}()
	}
}

// pendingTransactionsPage redirects the user to the dashboard (/) page after the user login.
func (handler *transferHandler) pendingTransactionsPage() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/#transactions", http.StatusFound)
	}
}
