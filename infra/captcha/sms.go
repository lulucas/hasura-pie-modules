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
	SmsCaptchaTTL = 5 * time.Minute
	SendInterval  = time.Minute
)

type Sms interface {
	SendCaptcha(ctx context.Context, mobile, captcha string, ttl time.Duration) error
}

func sendSmsCaptcha(cc pie.CreatedContext) interface{} {
	type SendSmsCaptchaOutput struct {
		Result bool
	}
	sms := cc.Get("sms").(Sms)
	r := cc.Get("redis").(*redis.Client)
	return func(ctx context.Context, input struct {
		Mobile string
	}) (*SendSmsCaptchaOutput, error) {
		key := fmt.Sprintf("captcha:sms:%s", input.Mobile)
		// check send interval
		ttl, err := r.TTL(key).Result()
		if err != nil {
			return nil, err
		}
		if SmsCaptchaTTL-ttl < SendInterval {
			return nil, ErrSendSmsCaptchaTooQuick
		}

		// send sms
		code := rand.Intn(1_000_000)
		if err := sms.SendCaptcha(ctx, input.Mobile, fmt.Sprintf("%06d", code), SmsCaptchaTTL); err != nil {
			return nil, err
		}
		// record captcha code to redis
		if err := r.Set(key, code, SmsCaptchaTTL).Err(); err != nil {
			return nil, err
		}
		cc.Logger().Infof("Send sms captcha %06d to %s", code, input.Mobile)

		return &SendSmsCaptchaOutput{Result: true}, nil
	}
}

func (m *captcha) ValidateSmsCaptcha(mobile, captcha string) (err error) {
	key := fmt.Sprintf("captcha:sms:%s", mobile)

	// compare result
	record, err := m.r.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrInvalidSmsCaptcha
		}
		return err
	}

	if record != captcha {
		return ErrInvalidSmsCaptcha
	}

	if redisErr := m.r.Del(fmt.Sprintf("captcha:sms:%s", mobile)).Err(); redisErr != nil {
		err = redisErr
	}

	m.logger.Infof("Validate sms captcha %s to %s success", captcha, mobile)

	return nil
}
