package model

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type PayLog struct {
	// 编号
	Id uuid.UUID
	// 支付通道号
	PayChannelId int32
	// 订单编号
	OrderId uuid.UUID
	// 用户编号
	UserId *uuid.UUID
	// 订单金额
	OrderAmount decimal.Decimal `pg:",use_zero"`
	// 实收金额
	ReceivedAmount decimal.Decimal `pg:",use_zero"`
	// 外部订单号
	OutTradeNo string
	// 附加信息
	Attach string
	// 是否支付
	IsPaid bool
}
