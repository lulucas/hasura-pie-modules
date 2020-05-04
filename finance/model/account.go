package model

import uuid "github.com/satori/go.uuid"

type WithdrawAccount struct {
	Id       uuid.UUID
	UserId   uuid.UUID
	Bank     string
	Identity string
	Holder   string
	QrCode   *string
	Enabled  bool
	Priority int
}
