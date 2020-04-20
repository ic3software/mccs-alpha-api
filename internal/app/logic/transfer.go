package logic

import (
	"errors"
	"fmt"
	"math"

	"github.com/ic3network/mccs-alpha-api/internal/app/repository/es"
	"github.com/ic3network/mccs-alpha-api/internal/app/repository/pg"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

type transfer struct{}

var Transfer = &transfer{}

func (t *transfer) Search(req *types.SearchTransferReqBody) (*types.SearchTransferRespond, error) {
	transfers, err := pg.Journal.Search(req)
	if err != nil {
		return nil, err
	}
	return transfers, nil
}

func (t *transfer) FindJournal(transferID string) (*types.Journal, error) {
	journal, err := pg.Journal.FindJournal(transferID)
	if err != nil {
		return nil, err
	}
	return journal, nil
}

func (t *transfer) Create(req *types.TransferReqBody) (*types.Journal, error) {
	journal, err := pg.Journal.Create(req)
	if err != nil {
		return nil, err
	}
	err = es.Journal.Create(journal)
	if err != nil {
		return nil, err
	}
	return journal, nil
}

func (t *transfer) Propose(req *types.TransferReqBody) (*types.Journal, error) {
	journal, err := pg.Journal.Propose(req)
	if err != nil {
		return nil, err
	}
	err = es.Journal.Create(journal)
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

	exceed, err := BalanceLimit.IsExceedLimit(from.AccountNumber, from.Balance-req.Amount)
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

	exceed, err = BalanceLimit.IsExceedLimit(to.AccountNumber, to.Balance+req.Amount)
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
	maxPosBal, err := BalanceLimit.GetMaxPosBalance(a.AccountNumber)
	if err != nil {
		return 0, err
	}
	if a.Balance >= 0 {
		return maxPosBal - a.Balance, nil
	}
	return math.Abs(a.Balance) + maxPosBal, nil
}

func (t *transfer) maxNegativeBalanceCanBeTransferred(a *types.Account) (float64, error) {
	maxNegBal, err := BalanceLimit.GetMaxNegBalance(a.AccountNumber)
	if err != nil {
		return 0, err
	}
	if a.Balance >= 0 {
		return a.Balance + maxNegBal, nil
	}
	return maxNegBal - math.Abs(a.Balance), nil
}

func (t *transfer) Cancel(transferID string, reason string) (*types.Journal, error) {
	j, err := pg.Journal.Cancel(transferID, reason)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (t *transfer) Accept(j *types.Journal) (*types.Journal, error) {
	updated, err := pg.Journal.Accept(j)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// GET /admin/transfers/{transferID}

func (t *transfer) AdminGetTransfer(transferID string) (*types.Journal, error) {
	journal, err := pg.Journal.FindJournal(transferID)
	if err != nil {
		return nil, err
	}
	return journal, nil
}
