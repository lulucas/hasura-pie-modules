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
	OrderAmount    decimal.Decimal
	ReceivedAmount decimal.Decimal
	OrderId        uuid.UUID
	OutTradeNo     string
	Attach         string
	IsPaid         bool
}
