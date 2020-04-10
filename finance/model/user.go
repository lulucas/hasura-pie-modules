package model

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	Id            uuid.UUID
	Balance       decimal.Decimal `pg:",use_zero"`
	FrozenBalance decimal.Decimal `pg:",use_zero"`
}
