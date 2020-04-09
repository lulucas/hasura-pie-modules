package captcha

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-redis/redis/v7"
	pie "github.com/lulucas/hasura-pie"
	"github.com/xkeyideal/captcha/pool"
	"time"
)

const (
	ImageCaptchaTTL = 5 * time.Minute
)

type CreateImageCaptchaOutput struct {
	Id    string
	Image string
}

func createImageCaptcha(cc pie.CreatedContext) interface{} {
	r := cc.Get("redis").(*redis.Client)

	captchaPool := pool.NewCaptchaPool(240, 80, 6, 10, 1, 2)

	return func(ctx context.Context) (*CreateImageCaptchaOutput, error) {
		img := captchaPool.GetImage()
		if err := r.Set(fmt.Sprintf("captcha:image:%s", img.Id), img.Val, ImageCaptchaTTL).Err(); err != nil {
			return nil, err
		}
		return &CreateImageCaptchaOutput{
			Id:    img.Id,
			Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(img.Data.Bytes()),
		}, nil
	}
}

func (m *captcha) ValidateImageCaptcha(id, captcha string) (err error) {
	origin, err := m.r.Get(fmt.Sprintf("captcha:image:%s", id)).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrorImageCaptchaInvalid
		}
		return err
	}

	defer func() {
		if redisErr := m.r.Del(fmt.Sprintf("captcha:image:%s", id)).Err(); redisErr != nil {
			err = redisErr
		}
	}()

	if origin != captcha {
		return ErrorImageCaptchaInvalid
	}

	m.logger.Infof("Validate image captcha %s to %s success", captcha, id)

	return nil
}
