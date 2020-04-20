package pg

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
)

type journal struct{}

var Journal = &journal{}

func (t *journal) Search(req *types.SearchTransferReqBody) (*types.SearchTransferRespond, error) {
	var journals []*types.Journal
	var numberOfResults int
	var err error

	countSQL := `
		SELECT COUNT(*)
		FROM journals
		WHERE (from_account_number = ? OR to_account_number = ?)
	`
	searchSQL := `
		SELECT *
		FROM journals
		WHERE (from_account_number = ? OR to_account_number = ?)
	`

	if req.Status == "all" {
		err = db.Raw(countSQL, req.QueryingAccountNumber, req.QueryingAccountNumber).Count(&numberOfResults).Error
		err = db.Raw(searchSQL+"ORDER BY created_at DESC LIMIT ? OFFSET ?",
			req.QueryingAccountNumber,
			req.QueryingAccountNumber,
			req.PageSize,
			req.Offset).
			Scan(&journals).Error
	} else {
		err = db.Raw(countSQL+"AND status = ?", req.QueryingAccountNumber, req.QueryingAccountNumber, constant.MapTransferType(req.Status)).
			Count(&numberOfResults).Error
		err = db.Raw(searchSQL+"AND status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
			req.QueryingAccountNumber,
			req.QueryingAccountNumber,
			constant.MapTransferType(req.Status),
			req.PageSize,
			req.Offset).
			Scan(&journals).Error
	}
	if err != nil {
		return nil, err
	}

	found := &types.SearchTransferRespond{
		Transfers:       types.JournalsToTransfers(journals, req.QueryingAccountNumber),
		NumberOfResults: numberOfResults,
		TotalPages:      util.GetNumberOfPages(numberOfResults, req.PageSize),
	}

	return found, nil
}

func (t *journal) Create(req *types.TransferReqBody) (*types.Journal, error) {
	tx := db.Begin()
	journal, err := t.propose(tx, req)
	if err != nil {
		return nil, err
	}
	updated, err := t.accept(tx, journal)
	if err != nil {
		return nil, err
	}
	return updated, err
}

func (t *journal) Propose(req *types.TransferReqBody) (*types.Journal, error) {
	tx := db.Begin()
	return t.propose(tx, req)
}

func (t *journal) propose(tx *gorm.DB, req *types.TransferReqBody) (*types.Journal, error) {
	journalRecord := &types.Journal{
		TransferID:        ksuid.New().String(),
		InitiatedBy:       req.InitiatorAccountNumber,
		FromAccountNumber: req.FromAccountNumber,
		FromEmail:         req.FromEmail,
		FromEntityName:    req.FromEntityName,
		ToAccountNumber:   req.ToAccountNumber,
		ToEmail:           req.ToEmail,
		ToEntityName:      req.ToEntityName,
		Amount:            req.Amount,
		Description:       req.Description,
		Type:              constant.Journal.Transfer,
		Status:            constant.Transfer.Initiated,
	}
	err := tx.Create(journalRecord).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return journalRecord, nil
}

func (t *journal) FindJournal(transferID string) (*types.Journal, error) {
	var result types.Journal

	err := db.Raw(`
		SELECT *
		FROM journals
		WHERE transfer_id = ?
		LIMIT 1
	`, transferID).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *journal) Cancel(transferID string, reason string) (*types.Journal, error) {
	err := db.Exec(`
		UPDATE journals
		SET status = ?, cancellation_reason = ?, updated_at = ?
		WHERE transfer_id = ?
	`, constant.Transfer.Cancelled, reason, time.Now(), transferID).Error
	if err != nil {
		return nil, err
	}

	var updated types.Journal
	err = db.Raw(`
		SELECT *
		FROM journals
		WHERE transfer_id=?
	`, transferID).Scan(&updated).Error
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (t *journal) Accept(j *types.Journal) (*types.Journal, error) {
	tx := db.Begin()
	return t.accept(tx, j)
}

func (t *journal) accept(tx *gorm.DB, j *types.Journal) (*types.Journal, error) {
	// Create postings.
	err := tx.Create(&types.Posting{
		AccountNumber: j.FromAccountNumber,
		JournalID:     j.ID,
		Amount:        -j.Amount,
	}).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Create(&types.Posting{
		AccountNumber: j.ToAccountNumber,
		JournalID:     j.ID,
		Amount:        j.Amount,
	}).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update accounts' balance.
	err = tx.Model(&types.Account{}).Where("account_number = ?", j.FromAccountNumber).Update("balance", gorm.Expr("balance - ?", j.Amount)).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Model(&types.Account{}).Where("account_number = ?", j.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", j.Amount)).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update the transaction status.
	err = tx.Exec(`
		UPDATE journals
		SET status = ?, completed_at = ?, updated_at = ?
		WHERE transfer_id=?
		RETURNING *
	`, constant.Transfer.Completed, time.Now(), time.Now(), j.TransferID).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var updated types.Journal
	err = tx.Raw(`
		SELECT *
		FROM journals
		WHERE transfer_id=?
	`, j.TransferID).Scan(&updated).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &updated, tx.Commit().Error
}

// TO BE REMOVED

// // FindPendings finds the pending transactions.
// func (journal *transfer) FindPendings(id uint) ([]*types.SearchTransferRespond, error) {
// 	var result []*types.SearchTransferRespond
// 	err := db.Raw(`
// 	SELECT
// 		J.id, J.transaction_id, CAST((CASE WHEN J.initiated_by = ? THEN 1 ELSE 0 END) AS BIT) AS "is_initiator",
// 		J.id, J.initiated_by, J.from_id, J.from_email, J.to_id, J.from_entity_name, J.to_entity_name,
// 		J.to_email, J.amount, J.description, J.created_at
// 	FROM journals AS J
// 	WHERE (J.from_id = ? OR J.to_id = ?) AND J.status = ?
// 	ORDER BY J.created_at DESC
// 	`, id, id, id, constant.Transfer.Initiated).Scan(&result).Error

// 	if err != nil {
// 		return nil, e.Wrap(err, "pg.Transaction.FindPendingTransactions failed")
// 	}
// 	return result, nil
// }

// // FindRecent finds the recent 3 completed transactions.
// func (journal *transfer) FindRecent(id uint) ([]*types.SearchTransferRespond, error) {
// 	var result []*types.SearchTransferRespond
// 	err := db.Raw(`
// 	SELECT J.transaction_id, J.from_email, J.to_email, J.from_entity_name, J.to_entity_name, J.description, P.amount, P.created_at
// 	FROM postings AS P
// 	INNER JOIN journals AS J ON J."id" = P."journal_id"
// 	WHERE P.account_id = ?
// 	ORDER BY P.created_at DESC
// 	LIMIT ?
// 	`, id, 3).Scan(&result).Error

// 	if err != nil {
// 		return nil, e.Wrap(err, "pg.Transaction.FindRecent failed")
// 	}
// 	return result, nil
// }

// // FindInRange finds the completed transactions in specific time range.
// func (journal *transfer) FindInRange(id uint, dateFrom time.Time, dateTo time.Time, page int) ([]*types.SearchTransferRespond, int, error) {
// 	limit := viper.GetInt("page_size")
// 	offset := viper.GetInt("page_size") * (page - 1)

// 	if dateFrom.IsZero() {
// 		dateFrom = constant.Date.DefaultFrom
// 	}
// 	if dateTo.IsZero() {
// 		dateTo = constant.Date.DefaultTo
// 	}

// 	// Add 24 hours to include the end date.
// 	dateTo = dateTo.Add(24 * time.Hour)

// 	var result []*types.SearchTransferRespond
// 	err := db.Raw(`
// 	SELECT J.transaction_id, J.from_email, J.to_email, J.from_entity_name, J.to_entity_name, J.description, P.amount, P.created_at
// 	FROM postings AS P
// 	INNER JOIN journals AS J ON J."id" = P."journal_id"
// 	WHERE P.account_id = ? AND (P.created_at BETWEEN ? AND ?)
// 	ORDER BY P.created_at DESC
// 	LIMIT ? OFFSET ?
// 	`, id, dateFrom, dateTo, limit, offset).Scan(&result).Error

// 	var numberOfResults int64
// 	db.Model(&types.Posting{}).Where("account_id = ? AND (created_at BETWEEN ? AND ?)", id, dateFrom, dateTo).Count(&numberOfResults)
// 	totalPages := util.GetNumberOfPages(int(numberOfResults), viper.GetInt("page_size"))

// 	if err != nil {
// 		return nil, 0, e.Wrap(err, "pg.Transaction.Find failed")
// 	}
// 	return result, totalPages, nil
// }
