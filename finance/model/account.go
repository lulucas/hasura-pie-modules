package model

import uuid "github.com/satori/go.uuid"

type Account struct {
	Id              uuid.UUID
	UserId          uuid.UUID
	Bank            string
	Identity        string
	Holder          string
	QrCode          *string
	Enabled         bool
	WithdrawDefault bool
}
