package model

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"time"
)

type RechargeStatus string

const (
	RechargeStatusPending   RechargeStatus = "pending"
	RechargeStatusPassed    RechargeStatus = "passed"
	RechargeStatusCancelled RechargeStatus = "cancelled"
)

type RechargeLog struct {
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
	Status       RechargeStatus
	ClientIp     string
	ClientRegion string
}

const RechargeSettingKey = "finance.recharge"

type RechargeSetting struct {
}
