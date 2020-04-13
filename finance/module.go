package finance

import (
	"github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/finance/model"
	"github.com/shopspring/decimal"
)

type finance struct {
	opt option
}

type option struct {
	WithdrawEnabled bool `envconfig:"default=false"`
	RechargeEnabled bool `envconfig:"default=false"`
}

func New() *finance {
	return &finance{}
}

func (m *finance) BeforeCreated(bc pie.BeforeCreatedContext) {
	bc.LoadFromEnv(&m.opt)
	bc.InitConfig(model.WithdrawSettingKey, &model.WithdrawSetting{
		MinAmount:         decimal.NewFromInt(500),
		MaxAmount:         decimal.NewFromInt(5000),
		MinBalanceReserve: decimal.Zero,
	})
}

func (m *finance) Created(cc pie.CreatedContext) {
	if m.opt.WithdrawEnabled {
		cc.HandleAction("withdraw", withdraw(cc))
		cc.HandleAction("audit_withdraw", auditWithdraw(cc))
		cc.HandleAction("reject_withdraw", rejectWithdraw(cc))
	}
	if m.opt.RechargeEnabled {
		cc.HandleAction("recharge", recharge(cc))
		cc.HandleEvent("recharge_paid", rechargePaid(cc))
	}
}
