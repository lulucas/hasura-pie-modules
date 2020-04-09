package captcha

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v7"
	pie "github.com/lulucas/hasura-pie"
	"math/rand"
	"time"
)

const (
	// 验证码有效时间
	SmsCaptchaTTL = 5 * time.Minute
	// 发送间隔
	SendInterval = time.Minute
)

type Sms interface {
	SendCaptcha(ctx context.Context, mobile, captcha string, ttl time.Duration) error
}

type SendSmsCaptchaInput struct {
	Mobile string
}

type SendSmsCaptchaOutput struct {
	Result bool
}

func sendSmsCaptcha(cc pie.CreatedContext) interface{} {
	sms := cc.Get("sms").(Sms)
	r := cc.Get("redis").(*redis.Client)
	return func(ctx context.Context, input SendSmsCaptchaInput) (*SendSmsCaptchaOutput, error) {
		key := fmt.Sprintf("captcha:sms:%s", input.Mobile)
		// 检查发送间隔
		ttl, err := r.TTL(key).Result()
		if err != nil {
			return nil, err
		}
		if SmsCaptchaTTL-ttl < SendInterval {
			return nil, ErrorSendSmsCaptchaTooQuick
		}

		// 发送短信
		code := rand.Intn(1_000_000)
		if err := sms.SendCaptcha(ctx, input.Mobile, fmt.Sprintf("%06d", code), SmsCaptchaTTL); err != nil {
			return nil, err
		}
		// 写入redis记录
		if err := r.Set(key, code, SmsCaptchaTTL).Err(); err != nil {
			return nil, err
		}
		cc.Logger().Infof("Send sms captcha %06d to %s", code, input.Mobile)

		return &SendSmsCaptchaOutput{Result: true}, nil
	}
}

func (m *captcha) ValidateSmsCaptcha(mobile, captcha string) (err error) {
	key := fmt.Sprintf("captcha:sms:%s", mobile)

	// 比对redis记录
	record, err := m.r.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	defer func() {
		// 不论是否验证成功都删除验证码
		if redisErr := m.r.Del(fmt.Sprintf("captcha:sms:%s", mobile)).Err(); redisErr != nil {
			err = redisErr
		}
	}()

	if record != captcha {
		return nil
	}

	m.logger.Infof("Validate sms captcha %s to %s success", captcha, mobile)

	return nil
}
