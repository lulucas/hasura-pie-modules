package model

import uuid "github.com/satori/go.uuid"

type Account struct {
	Id              uuid.UUID
	UserId          uuid.UUID
	Account         string
	Bank            string
	Holder          string
	QrCode          *string
	Enabled         bool
	WithdrawDefault bool
}
