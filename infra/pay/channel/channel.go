package channel

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type Channel interface {
	Pay(orderId uuid.UUID, userId *uuid.UUID, amount decimal.Decimal, returnUrl, notifyUrl, title, clientIp string, channelId int32, params json.RawMessage) (method, data string, err error)
	Notify(c echo.Context, rawParams json.RawMessage) (*Notification, error)
	ConfirmNotify(c echo.Context) error
}

type Notification struct {
	// 订单金额
	OrderAmount decimal.Decimal
	// 实收金额
	ReceivedAmount decimal.Decimal
	// 订单编号
	OrderId uuid.UUID
	// 外部编号
	OutTradeNo string
	// 附加参数
	Attach string
	// 是否支付
	IsPaid bool
}
