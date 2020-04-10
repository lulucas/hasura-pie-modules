package model

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"time"
)

type WithdrawStatus string

const (
	WithdrawStatusPending  WithdrawStatus = "pending"
	WithdrawStatusPassed   WithdrawStatus = "passed"
	WithdrawStatusRejected WithdrawStatus = "rejected"
)

type WithdrawLog struct {
	Id           uuid.UUID
	UserId       uuid.UUID
	Amount       decimal.Decimal `pg:",use_zero"`
	Commission   decimal.Decimal `pg:",use_zero"`
	Bank         string
	Account      string
	Holder       string
	RemarkSubmit *string
	AuditorId    *uuid.UUID
	AuditedAt    *time.Time
	RemarkAudit  *string
	Status       WithdrawStatus
	ClientIp     string
	ClientRegion string
}

const WithdrawSettingKey = "finance.withdraw"

type WithdrawSetting struct {
	MinAmount         decimal.Decimal
	MaxAmount         decimal.Decimal
	MinBalanceReserve decimal.Decimal
}
