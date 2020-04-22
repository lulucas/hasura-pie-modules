package finance

import (
	"context"
	"errors"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/finance/model"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

func recharge(cc pie.CreatedContext) interface{} {
	type RechargeOutput struct {
		Id uuid.UUID
	}
	return func(ctx context.Context, input struct {
		AccountId *uuid.UUID
		Amount    decimal.Decimal
	}) (*RechargeOutput, error) {
		tx, err := cc.DB().WithContext(ctx).Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		config := model.RechargeConfig{}
		if err := cc.LoadConfig(&config); err != nil {
			return nil, err
		}

		userId := cc.GetSession(ctx).UserId

		user := model.User{}
		if err := tx.Model(&user).Where("id = ?", userId).Select(); err != nil {
			return nil, err
		}

		account := model.Account{}
		if input.AccountId != nil {
			if err := tx.Model(&account).Where("id = ?", input.AccountId).Where("enabled = ?", true).Select(); err != nil {
				return nil, err
			}
		}

		rechargeLog := model.RechargeLog{
			UserId:  *userId,
			Amount:  input.Amount,
			Bank:    account.Bank,
			Account: account.Identity,
			Holder:  account.Holder,
			Status:  model.RechargeStatusPending,
		}
		if _, err := tx.Model(&rechargeLog).Insert(); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}

		cc.Logger().Infof("Recharge order %s to %s created", input.Amount, userId)

		return &RechargeOutput{
			Id: rechargeLog.Id,
		}, nil
	}
}

func rechargePaid(cc pie.CreatedContext) interface{} {
	type Event struct {
		pie.Event
		Old model.PayLog
		New model.PayLog
	}
	return func(ctx context.Context, evt Event) error {
		tx, err := cc.DB().WithContext(ctx).Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		recharged := !evt.Old.IsPaid && evt.New.IsPaid

		status := model.RechargeStatusCancelled
		if recharged {
			status = model.RechargeStatusPassed
		}

		rechargeLog := model.RechargeLog{}
		if _, err := tx.Model(&rechargeLog).
			Where("id = ?", evt.New.OrderId).
			Where("status = ?", model.RechargeStatusPending).
			Set("status = ?", status).
			Update(); err != nil {
			return err
		}

		if recharged {
			user := model.User{}
			updated, err := tx.Model(&user).
				Where("id = ?", evt.New.UserId).
				Set("balance = balance + ?", rechargeLog.Amount).
				Update()
			if err != nil {
				return err
			}
			if updated.RowsAffected() != 1 {
				return errors.New("finance.recharge.user-update-failed")
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		cc.Logger().Infof("Recharge order %s to %s confirmed", rechargeLog.Amount, evt.New.UserId)

		return nil
	}
}
