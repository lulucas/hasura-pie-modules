package sms

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/infra/sms/channel"
	"github.com/lulucas/hasura-pie-modules/infra/sms/channel/fake"
	"github.com/lulucas/hasura-pie-modules/infra/sms/channel/jiekou"
	"github.com/lulucas/hasura-pie-modules/infra/sms/channel/qcloud"
	"github.com/lulucas/hasura-pie-modules/infra/sms/model"
	"github.com/sarulabs/di"
	"time"
)

var (
	ErrChannelNotFound = errors.New("sms.channel.not-found")
)

type sms struct {
	channels map[string]channel.Channel
	db       *pg.DB
	logger   pie.Logger
}

func New() *sms {
	return &sms{
		channels: map[string]channel.Channel{
			"fake":   fake.New(),
			"jiekou": jiekou.New(),
			"qcloud": qcloud.New(),
		},
	}
}

func (m *sms) BeforeCreated(bc pie.BeforeCreatedContext) {
	bc.Add(di.Def{
		Name: "sms",
		Build: func(ctn di.Container) (i interface{}, err error) {
			return m, nil
		},
	})
}

func (m *sms) Created(cc pie.CreatedContext) {
	m.db = cc.DB()
	m.logger = cc.Logger()
}

func (m *sms) SendCaptchaByChannel(platform string, params json.RawMessage, mobile, captcha string, ttl time.Duration) error {
	if ch, ok := m.channels[platform]; ok {
		m.logger.Infof("Send sms captcha %s to %s by %s", captcha, mobile, platform)
		return ch.SendCaptcha(params, mobile, captcha, ttl)
	} else {
		return ErrChannelNotFound
	}
}

func (m *sms) SendCaptcha(ctx context.Context, mobile, captcha string, ttl time.Duration) error {
	ch := model.SmsChannel{}
	if err := m.db.WithContext(ctx).
		Model(&ch).Limit(1).Where("enabled = ?", true).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrChannelNotFound
		}
		return err
	}

	return m.SendCaptchaByChannel(ch.Platform, ch.Params, mobile, captcha, ttl)
}
