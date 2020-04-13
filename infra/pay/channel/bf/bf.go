package bf

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
	"strings"
)

const baseUrl = "http://bf.doeal.top/api/pay"

type BF struct {
}

type Params struct {
	MchId string `json:"mch_id" schema:"mch_id"`
	Key   string `json:"key" schema:"key"`
	Type  string `json:"type" schema:"type"`
}

type H5Request struct {
	MchId      string `json:"mch_id" schema:"mch_id"`
	Sign       string `json:"sign" schema:"sign"`
	Type       string `json:"type" schema:"type"`
	NotifyUrl  string `json:"notify_url" schema:"notify_url"`
	BackUrl    string `json:"back_url" schema:"back_url"`
	CardType   string `json:"card_type" schema:"card_type"`
	OutTradeNo string `json:"out_trade_no" schema:"out_trade_no"`
	Body       string `json:"body" schema:"body"`
	TotalFee   string `json:"total_fee" schema:"total_fee"`
	ClientIp   string `json:"client_ip" schema:"client_ip"`
	CardNo     string `json:"card_no" schema:"card_no"`
}

type H5Response struct {
	ErrorCode  string `json:"error_code" schema:"error_code"`
	ErrorMsg   string `json:"error_msg" schema:"error_msg"`
	Sign       string `json:"sign" schema:"sign"`
	OutTradeNo string `json:"out_trade_no" schema:"out_trade_no"`
	OrderId    string `json:"order_id" schema:"order_id"`
	TotalFee   string `json:"total_fee" schema:"total_fee"`
	PayUrl     string `json:"pay_url" schema:"pay_url"`
}

type Notify struct {
	ErrorCode  string `json:"error_code" schema:"error_code"`
	ErrorMsg   string `json:"error_msg" schema:"error_msg"`
	Sign       string `json:"sign" schema:"sign"`
	OutTradeNo string `json:"out_trade_no" schema:"out_trade_no"`
	OrderId    string `json:"order_id" schema:"order_id"`
	// pay_status
	// 0：unpaid
	// 1：success
	// 2：failed
	PayStatus string `json:"pay_status" schema:"pay_status"`
	TotalFee  string `json:"total_fee" schema:"total_fee"`
	Body      string `json:"body" schema:"body"`
}

func New() *BF {
	return &BF{}
}

func (ch *BF) Pay(section string, channelId int32, orderId uuid.UUID, userId *uuid.UUID, amount decimal.Decimal, title string, returnUrl string, notifyUrl string, attach string, clientIp string, rawParams json.RawMessage) (method string, data string, err error) {
	params := Params{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return "", "", err
	}
	signFunc := utils.NewSignFunc(utils.SignOption{
		IgnoreKeys:  []string{"error_code", "error_msg", "sign"},
		IgnoreEmpty: true,
		PostSignStrHook: func(s string) string {
			return s + params.Key
		},
	})
	req := H5Request{
		MchId:      params.MchId,
		BackUrl:    returnUrl,
		NotifyUrl:  utils.JoinNotifyUrl(section, notifyUrl, channelId, userId),
		CardType:   "2",
		OutTradeNo: strings.ReplaceAll(orderId.String(), "-", ""),
		Body:       title,
		TotalFee:   amount.StringFixed(2),
		ClientIp:   clientIp,
		Type:       params.Type,
	}
	sign, err := signFunc(req)
	if err != nil {
		return "", "", err
	}
	req.Sign = sign
	resp, err := grequests.Post(baseUrl, &grequests.RequestOptions{
		JSON: req,
	})
	if err != nil {
		return "", "", err
	}
	res := H5Response{}
	if err := resp.JSON(&res); err != nil {
		return "", "", err
	}

	if res.ErrorCode != "0" {
		if res.ErrorCode == "1005" {
			return "", "", errors.New("pay.channel-busy")
		}
		return "", "", errors.Errorf("pay gateway error code: %s, message: %s", res.ErrorCode, res.ErrorMsg)
	}

	sign, err = signFunc(res)
	if err != nil {
		return "", "", err
	}
	if sign != res.Sign {
		return "", "", errors.New("pay.invalid-sign")
	}

	return model.MethodUrl, res.PayUrl, nil
}

func (ch *BF) Notify(c echo.Context, rawParams json.RawMessage) (*channel.Notification, error) {
	notify := Notify{}
	if err := c.Bind(&notify); err != nil {
		return nil, err
	}

	params := Params{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, err
	}

	signFunc := utils.NewSignFunc(utils.SignOption{
		IgnoreKeys:  []string{"error_code", "error_msg", "sign"},
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
		return nil, errors.New("invalid sign")
	}

	amount, err := decimal.NewFromString(notify.TotalFee)
	if err != nil {
		return nil, err
	}

	orderId := uuid.FromStringOrNil(notify.OutTradeNo)

	return &channel.Notification{
		OrderAmount:    amount,
		ReceivedAmount: amount,
		OrderId:        orderId,
		OutTradeNo:     notify.OrderId,
		Attach:         "",
		IsPaid:         notify.PayStatus == "1",
	}, nil
}

func (ch *BF) ConfirmNotify(c echo.Context) error {
	return c.String(http.StatusOK, "SUCCESS")
}
