package finance

import (
	"context"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/finance/model"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"time"
)

var (
	ErrWithdrawUnknown = errors.New("finance.withdraw.unknown")
)

func withdraw(cc pie.CreatedContext) interface{} {
	type WithdrawOutput struct {
		Id uuid.UUID
	}
	return func(ctx context.Context, input struct {
		AccountId uuid.UUID
		Amount    decimal.Decimal
	}) (*WithdrawOutput, error) {
		tx, err := cc.DB().WithContext(ctx).Begin()
		if err != nil {
			return nil, err
		}

		setting := model.WithdrawSetting{}
		if err := cc.LoadConfig(model.WithdrawSettingKey, &setting); err != nil {
			return nil, err
		}

		if input.Amount.LessThan(setting.MinAmount) {
			return nil, errors.Errorf("finance.withdraw.min-amount:%s", setting.MinAmount)
		}
		if input.Amount.GreaterThan(setting.MaxAmount) {
			return nil, errors.Errorf("finance.withdraw.max-amount:%s", setting.MaxAmount)
		}

		userId := cc.GetSession(ctx).UserId

		user := model.User{}
		if err := tx.Model(&user).Where("id = ?", userId).Select(); err != nil {
			return nil, err
		}

		if user.Balance.LessThan(input.Amount) {
			return nil, errors.New("finance.withdraw.not-enough-balance")
		}
		if user.Balance.LessThan(input.Amount.Add(setting.MinBalanceReserve)) {
			return nil, errors.Errorf("finance.withdraw.min-balance:%s", setting.MinBalanceReserve)
		}

		account := model.Account{}
		if err := tx.Model(&account).Where("id = ?", input.AccountId).Where("enabled = ?", true).Select(); err != nil {
			return nil, err
		}

		withdrawLog := model.WithdrawLog{
			UserId:  *userId,
			Amount:  input.Amount,
			Bank:    account.Bank,
			Account: account.Identity,
			Holder:  account.Holder,
			Status:  model.WithdrawStatusPending,
		}
		if _, err := tx.Model(&withdrawLog).Insert(); err != nil {
			return nil, err
		}

		return &WithdrawOutput{
			Id: withdrawLog.Id,
		}, nil
	}
}

func auditWithdraw(cc pie.CreatedContext) interface{} {
	type AuditWithdrawOutput struct {
		Id uuid.UUID
	}
	return func(ctx context.Context, input struct {
		Id uuid.UUID
	}) (*AuditWithdrawOutput, error) {
		tx, err := cc.DB().WithContext(ctx).Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		withdrawLog := model.WithdrawLog{Id: input.Id}
		if err := tx.Select(&withdrawLog); err != nil {
			return nil, err
		}

		// 冻结余额
		userId := cc.GetSession(ctx).UserId
		user := model.User{}
		res, err := tx.Model(&user).
			Set("balance_frozen = balance_frozen - ?", withdrawLog.Amount).
			Where("id = ?", withdrawLog.UserId).
			Where("balance_frozen >= ?", withdrawLog.Amount).
			Update()
		if err != nil {
			return nil, err
		}
		if res.RowsAffected() != 1 {
			return nil, ErrWithdrawUnknown
		}

		res, err = tx.Model(&withdrawLog).
			Set("status = ?", model.WithdrawStatusPassed).
			Set("audited_at = ?", time.Now()).
			Set("auditor_id = ?", userId).
			Where("status = ?", model.WithdrawStatusPending).
			WherePK().Update()
		if err != nil {
			return nil, err
		}
		if res.RowsAffected() != 1 {
			return nil, ErrWithdrawUnknown
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		cc.Logger().Infof("Auditor %s audits user %s withdraw %s", userId, withdrawLog.UserId, withdrawLog.Id)

		return &AuditWithdrawOutput{
			Id: input.Id,
		}, nil
	}
}

func rejectWithdraw(cc pie.CreatedContext) interface{} {
	type RejectWithdrawOutput struct {
		Id uuid.UUID
	}
	return func(ctx context.Context, input struct {
		Id uuid.UUID
	}) (*RejectWithdrawOutput, error) {
		tx, err := cc.DB().WithContext(ctx).Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		withdrawLog := model.WithdrawLog{Id: input.Id}
		if err := tx.Select(&withdrawLog); err != nil {
			return nil, err
		}

		userId := cc.GetSession(ctx).UserId

		user := model.User{}
		res, err := tx.Model(&user).
			Set("balance_frozen = balance_frozen + ?", withdrawLog.Amount).
			Where("id = ?", withdrawLog.UserId).
			Update()
		if err != nil {
			return nil, err
		}
		if res.RowsAffected() != 1 {
			return nil, ErrWithdrawUnknown
		}

		res, err = tx.Model(&withdrawLog).
			Set("status = ?", model.WithdrawStatusRejected).
			Set("audited_at = ?", time.Now()).
			Set("auditor_id = ?", userId).
			Where("id = ?", input.Id).
			Where("status = ?", model.WithdrawStatusPending).
			Update()
		if err != nil {
			return nil, err
		}
		if res.RowsAffected() != 1 {
			return nil, ErrWithdrawUnknown
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		cc.Logger().Infof("Auditor %s audits user %s withdraw %s", userId, withdrawLog.UserId, withdrawLog.Id)

		return &RejectWithdrawOutput{
			Id: input.Id,
		}, nil
	}
}
