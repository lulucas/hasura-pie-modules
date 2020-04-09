package captcha

import (
	"errors"
	"github.com/go-redis/redis/v7"
	pie "github.com/lulucas/hasura-pie"
	"github.com/sarulabs/di"
)

var (
	ErrorSendSmsCaptchaTooQuick = errors.New("captcha.sms.send-too-quick")
	ErrorImageCaptchaInvalid    = errors.New("captcha.image.captcha-invalid")
)

type captcha struct {
	sms    Sms
	r      *redis.Client
	logger pie.Logger
}

func (m *captcha) BeforeCreated(bc pie.BeforeCreatedContext) {
	bc.Add(di.Def{
		Name: "captcha",
		Build: func(ctn di.Container) (i interface{}, err error) {
			m.sms = ctn.Get("sms").(Sms)
			return m, nil
		},
	})
}

func (m *captcha) Created(cc pie.CreatedContext) {
	m.sms = cc.Get("sms").(Sms)
	m.r = cc.Get("redis").(*redis.Client)
	m.logger = cc.Logger()

	cc.HandleAction("send_sms_captcha", sendSmsCaptcha(cc))
	cc.HandleAction("create_image_captcha", createImageCaptcha(cc))
}

func New() *captcha {
	return &captcha{}
}
