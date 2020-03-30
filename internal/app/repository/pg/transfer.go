package pg

import (
	"time"

	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
)

type transfer struct{}

var Transfer = &transfer{}

func (t *transfer) Search(q *types.SearchTransferQuery) (*types.SearchTransferRespond, error) {
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

	if q.Status == "all" {
		err = db.Raw(countSQL, q.QueryingAccountNumber, q.QueryingAccountNumber).Count(&numberOfResults).Error
		err = db.Raw(searchSQL+"ORDER BY created_at DESC LIMIT ? OFFSET ?",
			q.QueryingAccountNumber,
			q.QueryingAccountNumber,
			q.PageSize,
			q.Offset).
			Scan(&journals).Error
	} else {
		err = db.Raw(countSQL+"AND status = ?", q.QueryingAccountNumber, q.QueryingAccountNumber, constant.MapTransferType(q.Status)).
			Count(&numberOfResults).Error
		err = db.Raw(searchSQL+"AND status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
			q.QueryingAccountNumber,
			q.QueryingAccountNumber,
			constant.MapTransferType(q.Status),
			q.PageSize,
			q.Offset).
			Scan(&journals).Error
	}
	if err != nil {
		return nil, err
	}

	found := &types.SearchTransferRespond{
		Transfers:       t.journalsToTransfers(journals, q.QueryingAccountNumber),
		NumberOfResults: numberOfResults,
		TotalPages:      util.GetNumberOfPages(numberOfResults, q.PageSize),
	}

	return found, nil
}

func (t *transfer) journalsToTransfers(journals []*types.Journal, queryingAccountNumber string) []*types.Transfer {
	transfers := []*types.Transfer{}

	for _, j := range journals {
		t := &types.Transfer{
			TransferID:  j.TransferID,
			Description: j.Description,
			Amount:      j.Amount,
			CreatedAt:   &j.CreatedAt,
			Status:      j.Status,
		}
		if j.InitiatedBy == queryingAccountNumber {
			t.IsInitiator = true
		}
		if j.FromAccountNumber == queryingAccountNumber {
			t.Transfer = "out"
			t.AccountNumber = j.ToAccountNumber
			t.EntityName = j.ToEntityName
		} else {
			t.Transfer = "in"
			t.AccountNumber = j.FromAccountNumber
			t.EntityName = j.FromEntityName
		}
		if j.Status == constant.Transfer.Completed {
			t.CompletedAt = &j.UpdatedAt
		}

		transfers = append(transfers, t)
	}

	return transfers
}

func (t *transfer) Propose(req *types.TransferReqBody) (*types.Journal, error) {
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
	err := db.Create(journalRecord).Error
	if err != nil {
		return nil, err
	}

	return journalRecord, nil
}

func (t *transfer) FindJournal(transferID string) (*types.Journal, error) {
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

func (t *transfer) Cancel(transferID string, reason string) (*types.Journal, error) {
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

func (t *transfer) Accept(j *types.Journal) (*types.Journal, error) {
	tx := db.Begin()

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

// Create makes a transaction directly.
// func (t *transfer) Create(
// 	fromID uint,
// 	fromEmail string,
// 	fromEntityName string,

// 	toID uint,
// 	toEmail string,
// 	toEntityName string,

// 	amount float64,
// 	desc string,
// ) error {
// 	tx := db.Begin()

// 	journalRecord := &types.Journal{
// 		TransferID:        ksuid.New().String(),
// 		FromAccountNumber: "TODO",
// 		FromEmail:         fromEmail,
// 		FromEntityName:    fromEntityName,
// 		ToAccountNumber:   "TODO",
// 		ToEmail:           toEmail,
// 		ToEntityName:      toEntityName,
// 		Amount:            amount,
// 		Description:       desc,
// 		Type:              constant.Journal.Transfer,
// 		Status:            constant.Transfer.Completed,
// 	}
// 	err := tx.Create(journalRecord).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return e.Wrap(err, "pg.Transaction.Create")
// 	}

// 	journalID := journalRecord.ID

// 	// Create postings.
// 	err = tx.Create(&types.Posting{AccountNumber: fromID, JournalID: journalID, Amount: -amount}).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return e.Wrap(err, "pg.Transaction.Create")
// 	}
// 	err = tx.Create(&types.Posting{AccountNumber: toID, JournalID: journalID, Amount: amount}).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return e.Wrap(err, "pg.Transaction.Create")
// 	}

// 	// Update accounts' balance.
// 	err = tx.Model(&types.Account{}).Where("id = ?", fromID).Update("balance", gorm.Expr("balance - ?", amount)).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return e.Wrap(err, "pg.Transaction.Create")
// 	}
// 	err = tx.Model(&types.Account{}).Where("id = ?", toID).Update("balance", gorm.Expr("balance + ?", amount)).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return e.Wrap(err, "pg.Transaction.Create")
// 	}

// 	return tx.Commit().Error
// }

// // FindPendings finds the pending transactions.
// func (t *transfer) FindPendings(id uint) ([]*types.SearchTransferRespond, error) {
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
// func (t *transfer) FindRecent(id uint) ([]*types.SearchTransferRespond, error) {
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
// func (t *transfer) FindInRange(id uint, dateFrom time.Time, dateTo time.Time, page int) ([]*types.SearchTransferRespond, int, error) {
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
