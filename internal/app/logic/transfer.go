package logic

import (
	"errors"
	"fmt"
	"math"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/e"
)

type transfer struct{}

var Transfer = &transfer{}

func (t *transfer) Search(q *types.SearchTransferQuery) (*types.SearchTransferRespond, error) {
	transactions, err := pg.Transfer.Search(q)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *transfer) Propose(proposal *types.TransferProposal) (*types.Journal, error) {
	journal, err := pg.Transfer.Propose(proposal)
	if err != nil {
		return nil, err
	}
	return journal, nil
}

func (t *transfer) CheckBalance(proposal *types.TransferProposal) error {
	from, err := pg.Account.FindByAccountNumber(proposal.FromAccountNumber)
	if err != nil {
		return err
	}
	to, err := pg.Account.FindByAccountNumber(proposal.ToAccountNumber)
	if err != nil {
		return err
	}

	exceed, err := BalanceLimit.IsExceedLimit(from.ID, from.Balance-proposal.Amount)
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

	exceed, err = BalanceLimit.IsExceedLimit(to.ID, to.Balance+proposal.Amount)
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

func (t *transfer) Cancel(transactionID uint, reason string) error {
	err := pg.Transfer.Cancel(transactionID, reason)
	if err != nil {
		return err
	}
	return nil
}

func (t *transfer) Accept(
	transactionID uint,
	fromID uint,
	toID uint,
	amount float64,
) error {
	err := pg.Transfer.Accept(
		transactionID,
		fromID,
		toID,
		amount,
	)
	if err != nil {
		return e.Wrap(err, "service.Account.MakeTransfer failed")
	}
	return nil
}

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
