package model

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

type SmsChannel struct {
	// 编号
	Id uuid.UUID
	// 平台
	Platform string
	// 参数
	Params json.RawMessage
	// 是否启用
	Enabled bool
}
