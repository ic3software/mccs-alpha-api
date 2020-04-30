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

// POST /transfers

func (t *journal) Propose(req *types.TransferReqBody) (*types.Journal, error) {
	tx := db.Begin()
	journal, err := t.propose(tx, req)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return journal, tx.Commit().Error
}

func (t *journal) propose(tx *gorm.DB, req *types.TransferReqBody) (*types.Journal, error) {
	journalRecord := &types.Journal{
		TransferID:        ksuid.New().String(),
		InitiatedBy:       req.InitiatorAccountNumber,
		FromAccountNumber: req.FromAccountNumber,
		FromEntityName:    req.FromEntityName,
		ToAccountNumber:   req.ToAccountNumber,
		ToEntityName:      req.ToEntityName,
		Amount:            req.Amount,
		Description:       req.Description,
		Type:              req.TransferType,
		Status:            constant.Transfer.Initiated,
	}
	err := tx.Create(journalRecord).Error
	if err != nil {
		return nil, err
	}
	return journalRecord, nil
}

// GET /transfers

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
		Transfers:       types.NewJournalsToTransfersRespond(journals, req.QueryingAccountNumber),
		NumberOfResults: numberOfResults,
		TotalPages:      util.GetNumberOfPages(numberOfResults, req.PageSize),
	}

	return found, nil
}

func (t *journal) FindByID(transferID string) (*types.Journal, error) {
	var result types.Journal

	err := db.Raw(`
		SELECT *
		FROM journals
		WHERE deleted_at IS NULL AND transfer_id = ?
		LIMIT 1
	`, transferID).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// PATCH /transfers

func (t *journal) Cancel(transferID string, reason string) (*types.Journal, error) {
	err := db.Exec(`
		UPDATE journals
		SET status = ?, cancellation_reason = ?, updated_at = ?
		WHERE deleted_at IS NULL AND transfer_id = ?
	`, constant.Transfer.Cancelled, reason, time.Now(), transferID).Error
	if err != nil {
		return nil, err
	}

	var updated types.Journal
	err = db.Raw(`
		SELECT *
		FROM journals
		WHERE deleted_at IS NULL AND transfer_id = ?
	`, transferID).Scan(&updated).Error
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// PATCH /transfers

func (t *journal) Accept(j *types.Journal) (*types.Journal, error) {
	tx := db.Begin()
	journal, err := t.accept(tx, j)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return journal, tx.Commit().Error
}

func (t *journal) accept(tx *gorm.DB, j *types.Journal) (*types.Journal, error) {
	// Create postings.
	err := tx.Create(&types.Posting{
		AccountNumber: j.FromAccountNumber,
		JournalID:     j.ID,
		Amount:        -j.Amount,
	}).Error
	if err != nil {
		return nil, err
	}
	err = tx.Create(&types.Posting{
		AccountNumber: j.ToAccountNumber,
		JournalID:     j.ID,
		Amount:        j.Amount,
	}).Error
	if err != nil {
		return nil, err
	}

	// Update accounts' balance.
	err = tx.Model(&types.Account{}).Where("account_number = ?", j.FromAccountNumber).Update("balance", gorm.Expr("balance - ?", j.Amount)).Error
	if err != nil {
		return nil, err
	}
	err = tx.Model(&types.Account{}).Where("account_number = ?", j.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", j.Amount)).Error
	if err != nil {
		return nil, err
	}

	// Update the transaction status.
	err = tx.Exec(`
		UPDATE journals
		SET status = ?, completed_at = ?, updated_at = ?
		WHERE deleted_at IS NULL AND transfer_id = ?
		RETURNING *
	`, constant.Transfer.Completed, time.Now(), time.Now(), j.TransferID).Error
	if err != nil {
		return nil, err
	}

	var updated types.Journal
	err = tx.Raw(`
		SELECT *
		FROM journals
		WHERE transfer_id=?
	`, j.TransferID).Scan(&updated).Error
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// POST /admin/transfers

func (t *journal) Create(req *types.AdminTransferReqBody) (*types.Journal, error) {
	tx := db.Begin()
	journal, err := t.propose(tx, &types.TransferReqBody{
		FromAccountNumber: req.PayerEntity.AccountNumber,
		FromEntityName:    req.PayerEntity.EntityName,
		ToAccountNumber:   req.PayeeEntity.AccountNumber,
		ToEntityName:      req.PayeeEntity.EntityName,
		Amount:            req.Amount,
		Description:       req.Description,
		TransferType:      constant.TransferType.AdminTransfer,
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	updated, err := t.accept(tx, journal)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return updated, tx.Commit().Error
}

// GET /admin/transfers

func (t *journal) FindByIDs(transferIDs []string) ([]*types.Journal, error) {
	var journals []*types.Journal

	err := db.Where("transfer_id IN (?)", transferIDs).Find(&journals).Error
	if err != nil {
		return nil, err
	}

	return journals, nil
}

// GET /admin/entities/{entityID}

func (t *journal) AdminGetPendingTransfers(accountNumber string) ([]*types.AdminTransferRespond, error) {
	var journals []*types.Journal

	searchSQL := `
		SELECT *
		FROM journals
		WHERE deleted_at IS NULL AND (from_account_number = ? OR to_account_number = ?) AND status = ? ORDER BY created_at
	`
	err := db.Raw(searchSQL,
		accountNumber,
		accountNumber,
		constant.Transfer.Initiated,
	).Scan(&journals).Error
	if err != nil {
		return nil, err
	}

	return types.NewJournalsToAdminTransfersRespond(journals), nil
}
