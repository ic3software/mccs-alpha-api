package pg

import (
	"fmt"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

var BalanceLimit = &balanceLimit{}

type balanceLimit struct{}

func (b *balanceLimit) Create(tx *gorm.DB, accountNumber string) error {
	balance := &types.BalanceLimit{
		AccountNumber: accountNumber,
		MaxNegBal:     viper.GetFloat64("transaction.maxNegBal"),
		MaxPosBal:     viper.GetFloat64("transaction.maxPosBal"),
	}
	err := tx.Create(balance).Error
	if err != nil {
		return err
	}
	return nil
}

func (b *balanceLimit) FindByAccountNumber(accountNumber string) (*types.BalanceLimit, error) {
	var result types.BalanceLimit

	err := db.Raw(`
		SELECT max_pos_bal, max_neg_bal
		FROM balance_limits
		WHERE deleted_at IS NULL AND account_number = ?
		LIMIT 1
	`, accountNumber).Scan(&result).Error
	if err != nil {
		fmt.Println(accountNumber)
		return nil, err
	}

	return &result, nil
}

// PATCH /admin/entities/{entityID}

func (b *balanceLimit) AdminUpdate(req *types.AdminUpdateEntityReqBody) error {
	update := map[string]interface{}{}
	if req.MaxPosBal != nil {
		update["max_pos_bal"] = *req.MaxPosBal
	}
	if req.MaxNegBal != nil {
		update["max_neg_bal"] = *req.MaxNegBal
	}
	err := db.Table("balance_limits").Where("deleted_at IS NULL AND account_number = ?", req.OriginEntity.AccountNumber).Updates(update).Error
	if err != nil {
		return err
	}
	return nil
}

// DELETE /admin/entities/{entityID}

func (b *balanceLimit) Delete(accountNumber string) error {
	tx := db.Begin()

	err := tx.Exec(`
		UPDATE balance_limits
		SET deleted_at = ?, updated_at = ?
		WHERE deleted_at IS NULL AND account_number = ?
	`, time.Now(), time.Now(), accountNumber).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
