package mks101

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/levigross/grequests"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel"
	"github.com/lulucas/hasura-pie-modules/infra/pay/model"
	"github.com/lulucas/hasura-pie-modules/infra/pay/utils"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"net/http"
)

type Mks101 struct {
}

const baseUrl = "http://api.mks101.com/api/index/"

type Params struct {
	MerchantId string `json:"merchant_id" schema:"merchant_id"`
	Key        string `json:"key" schema:"key"`
	// 100	支付宝H5
	// 200	微信H5
	// 300	支付宝扫码
	// 400	微信扫码
	// 500	网银支付
	// 600	快捷支付
	// 700	网银对公
	// 800	银联扫码
	ProductId string `json:"product_id" schema:"product_id"`
}

type CreateOrderRequest struct {
	Amount     string `json:"amount" schema:"amount"`
	Attach     string `json:"attach" schema:"attach"`
	MerchantId string `json:"merchant_id" schema:"merchant_id"`
	NotifyUrl  string `json:"notify_url" schema:"notify_url"`
	OutTradeId string `json:"out_trade_id" schema:"out_trade_id"`
	ProductId  string `json:"product_id" schema:"product_id"`
	ReturnUrl  string `json:"return_url" schema:"return_url"`
	Sign       string `json:"sign" schema:"sign"`
}

type CreateOrderResponse struct {
	Code int    `json:"code" schema:"code"`
	Msg  string `json:"msg" schema:"msg"`
	Data struct {
		Url           string `json:"url" schema:"url"`
		TransactionId string `json:"transaction_id" schema:"transaction_id"`
	} `json:"data" schema:"data"`
}

type Notify struct {
	MerchantId string `json:"merchant_id" schema:"merchant_id"`
	Amount     string `json:"amount" schema:"amount"`
	Attach     string `json:"attach" schema:"attach"`
	OutTradeId string `json:"out_trade_id" schema:"out_trade_id"`
	ProductId  string `json:"product_id" schema:"product_id"`
	// 0: not paid; 1: paid
	Status        string `json:"status" schema:"status"`
	TransactionId string `json:"transaction_id" schema:"transaction_id"`
	Sign          string `json:"sign" schema:"sign"`
}

func New() *Mks101 {
	return &Mks101{}
}

func (ch *Mks101) Pay(section string, channelId int32, orderId uuid.UUID, userId *uuid.UUID, amount decimal.Decimal, title, returnUrl, notifyUrl, clientIp string, rawParams json.RawMessage) (method string, data string, err error) {
	params := Params{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return "", "", err
	}
	signFunc := utils.NewSignFunc(utils.SignOption{
		IgnoreKeys:  []string{"attach", "sign"},
		IgnoreEmpty: true,
		PostSignStrHook: func(s string) string {
			return s + "&key=" + params.Key
		},
	})
	req := CreateOrderRequest{
		Amount:     amount.StringFixed(2),
		Attach:     "",
		MerchantId: params.MerchantId,
		NotifyUrl:  utils.JoinNotifyUrl(section, notifyUrl, channelId, userId),
		OutTradeId: orderId.String(),
		ProductId:  params.ProductId,
		ReturnUrl:  returnUrl,
	}
	sign, err := signFunc(req)
	if err != nil {
		return "", "", err
	}
	req.Sign = sign
	resp, err := grequests.Post(baseUrl+"create_order", &grequests.RequestOptions{
		JSON: req,
	})
	if err != nil {
		return "", "", err
	}
	res := CreateOrderResponse{}
	if err := resp.JSON(&res); err != nil {
		return "", "", err
	}

	if res.Code != 1 {
		return "", "", errors.Errorf("pay gateway error code: %d, message: %s", res.Code, res.Msg)
	}

	return model.MethodUrl, res.Data.Url, nil
}

func (ch *Mks101) Notify(c echo.Context, rawParams json.RawMessage) (*channel.Notification, error) {
	notify := Notify{}
	if err := c.Bind(&notify); err != nil {
		return nil, err
	}

	params := Params{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, err
	}

	signFunc := utils.NewSignFunc(utils.SignOption{
		IgnoreKeys:  []string{"attach", "sign"},
		IgnoreEmpty: true,
		PostSignStrHook: func(s string) string {
			return s + params.Key
		},
	})

	sign, err := signFunc(notify)
	if err != nil {
		return nil, err
	}

	if notify.Sign != sign {
		return nil, errors.New("pay.mks101.invalid-sign")
	}

	amount, err := decimal.NewFromString(notify.Amount)
	if err != nil {
		return nil, err
	}

	orderId := uuid.FromStringOrNil(notify.OutTradeId)

	return &channel.Notification{
		OrderAmount:    amount,
		ReceivedAmount: amount,
		OrderId:        orderId,
		OutTradeNo:     notify.TransactionId,
		IsPaid:         notify.Status == "1",
	}, nil
}

func (ch *Mks101) ConfirmNotify(c echo.Context) error {
	return c.String(http.StatusOK, "success")
}
