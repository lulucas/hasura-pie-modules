package alipay

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel"
	"github.com/lulucas/hasura-pie-modules/infra/pay/model"
	"github.com/lulucas/hasura-pie-modules/infra/pay/utils"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/smartwalle/alipay/v3"
	"net/http"
)

type Alipay struct {
}

type Type string

const (
	TypePage   Type = "page"
	TypeQRCode Type = "qrcode"
)

type Params struct {
	IsProduction   bool   `json:"is_production"`
	AppId          string `json:"app_id"`
	Type           Type   `json:"type"`
	PrivateKey     string `json:"private_key"`
	AppCaPublicKey string `json:"app_ca_public_key"`
	RootCert       string `json:"root_cert"`
	CaPublicKey    string `json:"ca_public_key"`
}

func New() *Alipay {
	return &Alipay{}
}

func getClient(params Params) (*alipay.Client, error) {
	client, err := alipay.New(params.AppId, params.PrivateKey, params.IsProduction)
	if err != nil {
		return nil, err
	}
	if err := client.LoadAliPayRootCert(params.RootCert); err != nil {
		return nil, err
	}
	if err := client.LoadAliPayPublicCert(params.CaPublicKey); err != nil {
		return nil, err
	}
	if err := client.LoadAppPublicCert(params.AppCaPublicKey); err != nil {
		return nil, err
	}
	return client, nil
}

func (ch *Alipay) Pay(section string, channelId int32, orderId uuid.UUID, userId *uuid.UUID, amount decimal.Decimal, title, returnUrl, notifyUrl, clientIp string, rawParams json.RawMessage) (method string, data string, err error) {
	params := Params{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return "", "", err
	}

	client, err := getClient(params)
	if err != nil {
		return "", "", err
	}

	switch params.Type {
	case TypePage:
		pp := alipay.TradePagePay{}
		pp.NotifyURL = utils.JoinNotifyUrl(section, notifyUrl, channelId, userId)
		pp.ReturnURL = returnUrl
		pp.Subject = title
		pp.OutTradeNo = orderId.String()
		pp.TotalAmount = amount.StringFixed(2)
		pp.ProductCode = "FAST_INSTANT_TRADE_PAY"
		uri, err := client.TradePagePay(pp)
		if err != nil {
			return "", "", err
		}
		return model.MethodUrl, uri.String(), err
	case TypeQRCode:
		pp := alipay.TradePreCreate{}
		pp.NotifyURL = notifyUrl
		pp.ReturnURL = returnUrl
		pp.Subject = title
		pp.TotalAmount = amount.StringFixed(2)
		pp.OutTradeNo = orderId.String()
		rsp, err := client.TradePreCreate(pp)
		if err != nil {
			return "", "", err
		}
		if rsp.Content.Code != alipay.CodeSuccess {
			return "", "", errors.New(rsp.Content.Msg)
		}
		return model.MethodImage, rsp.Content.QRCode, err
	}
	return "", "", errors.Errorf("invalid type %s in alipay params", params.Type)
}

func (ch *Alipay) Notify(c echo.Context, rawParams json.RawMessage) (*channel.Notification, error) {
	params := Params{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, err
	}
	client, err := getClient(params)
	if err != nil {
		return nil, err
	}

	if err := c.Request().ParseForm(); err != nil {
		return nil, err
	}
	data := c.Request().PostForm

	if ok, err := client.VerifySign(data); err != nil {
		return nil, err
	} else {
		if !ok {
			return nil, errors.New("alipay invalid sign")
		}
	}

	orderAmount, err := decimal.NewFromString(data.Get("total_amount"))
	if err != nil {
		return nil, err
	}

	receivedAmount, err := decimal.NewFromString(data.Get("receipt_amount"))
	if err != nil {
		return nil, err
	}

	orderId, err := uuid.FromString(data.Get("out_trade_no"))
	if err != nil {
		return nil, err
	}

	return &channel.Notification{
		OrderAmount:    orderAmount,
		ReceivedAmount: receivedAmount,
		OrderId:        orderId,
		OutTradeNo:     data.Get("trade_no"),
		IsPaid:         data.Get("trade_status") == "TRADE_SUCCESS",
	}, nil

}

func (ch *Alipay) ConfirmNotify(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
