package model

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type RechargeStatus string

const (
	RechargeStatusPending   RechargeStatus = "pending"
	RechargeStatusFinished  RechargeStatus = "finished"
	RechargeStatusCancelled RechargeStatus = "cancelled"
)

type RechargeLog struct {
	Id           uuid.UUID
	UserId       uuid.UUID
	Amount       decimal.Decimal `pg:",use_zero"`
	Commission   decimal.Decimal `pg:",use_zero"`
	Status       RechargeStatus
	ClientIp     string
	ClientRegion string
}

type RechargeConfig struct {
}
