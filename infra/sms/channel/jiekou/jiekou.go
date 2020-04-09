package jiekou

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"time"
)

const baseUrl = "http://sms.106jiekou.com/utf8/sms.aspx"

type Jiekou struct {
}

func New() *Jiekou {
	return &Jiekou{}
}

func (j *Jiekou) SendTemplate(params Params, mobile string, content string) error {
	resp, err := grequests.Post(baseUrl, &grequests.RequestOptions{
		Params: map[string]string{
			"account":  params.Account,
			"password": params.Password,
			"mobile":   mobile,
			"content":  content,
		},
	})
	if err != nil {
		return err
	}

	code := resp.String()
	if code != "100" {
		return errors.New("jiekou error code " + code)
	}
	return nil
}
func (j *Jiekou) SendCaptcha(params json.RawMessage, mobile string, captcha string, ttl time.Duration) error {
	pm := Params{}
	if err := json.Unmarshal(params, &pm); err != nil {
		return err
	}
	return j.SendTemplate(pm, mobile, fmt.Sprintf("您的验证码是：%s。请不要把验证码泄露给其他人。如非本人操作，可不用理会！", captcha))
}

type Params struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}
