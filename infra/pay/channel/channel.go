package channel

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type Channel interface {
	Pay(orderId uuid.UUID, amount decimal.Decimal, subject, body, returnUrl, notifyUrl, clientIp string, params json.RawMessage) (method string, data string, err error)
	Notify(c echo.Context, params json.RawMessage) (*Notification, error)
	ConfirmNotify(c echo.Context) error
}

type Notification struct {
	OrderAmount    decimal.Decimal
	ReceivedAmount decimal.Decimal
	OrderId        uuid.UUID
	OutTradeNo     string
	IsPaid         bool
}
