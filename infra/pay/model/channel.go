package model

import "encoding/json"

type PayChannel struct {
	// 编号
	Id int32
	// 支付平台
	Platform string
	// 显示名称
	Title string
	// 默认前台返回地址
	ReturnUrl string
	// 默认后台通知地址
	NotifyUrl string
	// 参数
	Params json.RawMessage
	// 是否可见
	Visible bool
	// 是否可用
	Enabled bool
}
