package qcloud

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"math/rand"
	"strconv"
	"time"
)

type Qcloud struct {
}

func New() *Qcloud {
	return &Qcloud{}
}

type SendRequest struct {
	Ext    string   `json:"ext"`
	Extend string   `json:"extend"`
	Params []string `json:"params"`
	Sig    string   `json:"sig"`
	Sign   string   `json:"sign"`
	Tel    *Tel     `json:"tel"`
	Time   int64    `json:"time"`
	TplId  int64    `json:"tpl_id"`
}

type Tel struct {
	Mobile     string `json:"mobile"`
	NationCode string `json:"nationcode"`
}

type SmsResponse struct {
	Result int64  `json:"result"`
	ErrMsg string `json:"errmsg"`
	Ext    string `json:"ext"`
	Fee    int64  `json:"fee"`
	Sid    string `json:"sid"`
}

func (c *Qcloud) SendTemplate(params Params, templateId string, mobile string, args ...string) error {
	random := fmt.Sprintf("%010d", rand.Intn(10000000000))
	timeStamp := time.Now().Unix()
	h := sha256.New()
	format := fmt.Sprintf("appkey=%s&random=%s&time=%d&mobile=%s", params.AppKey, random, timeStamp, mobile)
	h.Write([]byte(format))

	tid, err := strconv.Atoi(templateId)
	if err != nil {
		return errors.New("invalid template id")
	}
	message := &SendRequest{
		Ext:    "",
		Extend: "",
		Params: args,
		Sig:    fmt.Sprintf("%x", h.Sum(nil)),
		Tel: &Tel{
			Mobile:     mobile,
			NationCode: "86",
		},
		Time:  timeStamp,
		TplId: int64(tid),
	}

	header := map[string]string{
		"Content-Type": "application/json",
	}

	ro := &grequests.RequestOptions{
		RequestTimeout: 5 * time.Second,
		Headers:        header,
		JSON:           message,
	}
	uri := fmt.Sprintf("https://yun.tim.qq.com/v5/tlssmssvr/sendsms?sdkappid=%s&random=%s", params.AppId, random)

	resp, err := grequests.Post(uri, ro)
	if err != nil {
		return err
	}

	result := &SmsResponse{}
	err = resp.JSON(result)
	if err != nil {
		return err
	}
	if result.ErrMsg != "OK" {
		return errors.New(fmt.Sprintf("send sms error, result code %d errmsg %s", result.Result, result.ErrMsg))
	}

	return nil
}

func (c *Qcloud) SendCaptcha(params json.RawMessage, mobile string, captcha string, ttl time.Duration) error {
	pm := Params{}
	if err := json.Unmarshal(params, &pm); err != nil {
		return err
	}
	return c.SendTemplate(pm, pm.CaptchaTplId, mobile, captcha, fmt.Sprintf("%d", int64(ttl.Minutes())))
}

type Params struct {
	AppId        string `json:"app_id"`
	AppKey       string `json:"app_key"`
	CaptchaTplId string `json:"captcha_tpl_id"`
}
