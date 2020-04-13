package model

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type PayLog struct {
	Id             uuid.UUID
	Section        string
	PayChannelId   int32
	OrderId        uuid.UUID
	UserId         *uuid.UUID
	OrderAmount    decimal.Decimal `pg:",use_zero"`
	ReceivedAmount decimal.Decimal `pg:",use_zero"`
	OutTradeNo     string
	Attach         string
	IsPaid         bool
}
