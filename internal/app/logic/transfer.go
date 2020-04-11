package logic

import (
	"errors"
	"fmt"
	"math"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type transfer struct{}

var Transfer = &transfer{}

func (t *transfer) Search(req *types.SearchTransferReqBody) (*types.SearchTransferRespond, error) {
	transactions, err := pg.Transfer.Search(req)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *transfer) FindJournal(transferID string) (*types.Journal, error) {
	journal, err := pg.Transfer.FindJournal(transferID)
	if err != nil {
		return nil, err
	}
	return journal, nil
}

func (t *transfer) Propose(req *types.TransferReqBody) (*types.Journal, error) {
	journal, err := pg.Transfer.Propose(req)
	if err != nil {
		return nil, err
	}
	return journal, nil
}

func (t *transfer) CheckBalance(req *types.TransferReqBody) error {
	from, err := pg.Account.FindByAccountNumber(req.FromAccountNumber)
	if err != nil {
		return err
	}
	to, err := pg.Account.FindByAccountNumber(req.ToAccountNumber)
	if err != nil {
		return err
	}

	exceed, err := BalanceLimit.IsExceedLimit(from.ID, from.Balance-req.Amount)
	if err != nil {
		return err
	}
	if exceed {
		amount, err := t.maxNegativeBalanceCanBeTransferred(from)
		if err != nil {
			return err
		}
		return errors.New("Sender will exceed its credit limit." + " The maximum amount that can be sent is: " + fmt.Sprintf("%.2f", amount))
	}

	exceed, err = BalanceLimit.IsExceedLimit(to.ID, to.Balance+req.Amount)
	if err != nil {
		return err
	}
	if exceed {
		amount, err := t.maxPositiveBalanceCanBeTransferred(to)
		if err != nil {
			return err
		}
		return errors.New("Receiver will exceed its maximum balance limit." + " The maximum amount that can be received is: " + fmt.Sprintf("%.2f", amount))
	}

	return nil
}

func (t *transfer) maxPositiveBalanceCanBeTransferred(a *types.Account) (float64, error) {
	maxPosBal, err := BalanceLimit.GetMaxPosBalance(a.ID)
	if err != nil {
		return 0, err
	}
	if a.Balance >= 0 {
		return maxPosBal - a.Balance, nil
	}
	return math.Abs(a.Balance) + maxPosBal, nil
}

func (t *transfer) maxNegativeBalanceCanBeTransferred(a *types.Account) (float64, error) {
	maxNegBal, err := BalanceLimit.GetMaxNegBalance(a.ID)
	if err != nil {
		return 0, err
	}
	if a.Balance >= 0 {
		return a.Balance + maxNegBal, nil
	}
	return maxNegBal - math.Abs(a.Balance), nil
}

func (t *transfer) Cancel(transferID string, reason string) (*types.Journal, error) {
	j, err := pg.Transfer.Cancel(transferID, reason)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (t *transfer) Accept(j *types.Journal) (*types.Journal, error) {
	updated, err := pg.Transfer.Accept(j)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// TO BE REMOVED

// func (t *transfer) Find(transactionID uint) (*types.Transfer, error) {
// 	transaction, err := pg.Transfer.Find(transactionID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return transaction, nil
// }

// func (t *transfer) FindPendings(accountID uint) ([]*types.Transfer, error) {
// 	transactions, err := pg.Transfer.FindPendings(accountID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return transactions, nil
// }

// func (t *transfer) FindRecent(accountID uint) ([]*types.Transfer, error) {
// 	transactions, err := pg.Transfer.FindRecent(accountID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return transactions, nil
// }

// func (t *transfer) FindInRange(accountID uint, dateFrom time.Time, dateTo time.Time, page int) ([]*types.Transfer, int, error) {
// 	transactions, totalPages, err := pg.Transfer.FindInRange(accountID, dateFrom, dateTo, page)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	return transactions, totalPages, nil
// }
