package account

import (
	"context"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Identifier       string
	Password         string
	Method           model.LoginMethod
	ImageCaptchaId   string
	ImageCaptchaCode string
}

type LoginOutput struct {
	Token string
}

func login(cc pie.CreatedContext, opt option) interface{} {
	c := cc.Get("captcha").(Captcha)

	return func(ctx context.Context, input LoginInput) (*LoginOutput, error) {
		if err := validation.ValidateStruct(&input,
			validation.Field(&input.Identifier, validation.Required, validation.Length(5, 32)),
			validation.Field(&input.Password, validation.Required, validation.Length(6, 32)),
		); err != nil {
			return nil, err
		}

		user := model.User{}

		switch input.Method {
		case model.LoginMethodName, model.LoginMethodMobile:
			if opt.LoginImageCaptcha {
				if err := c.ValidateImageCaptcha(input.ImageCaptchaId, input.ImageCaptchaCode); err != nil {
					return nil, err
				}
			}
			if err := validation.ValidateStruct(&input,
				validation.Field(&input.Password, validation.Required, validation.Length(6, 32)),
			); err != nil {
				return nil, err
			}
			if err := cc.DB().WithContext(ctx).Model(&user).Where(string(input.Method)+" = ?", input.Identifier).Select(); err != nil {
				if err == pg.ErrNoRows {
					return nil, ErrInvalidCredentials
				}
				return nil, err
			}
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
				return nil, ErrInvalidCredentials
			}
		case model.LoginMethodSms:
			if err := c.ValidateSmsCaptcha(input.Identifier, input.Password); err != nil {
				return nil, err
			}
			if err := cc.DB().WithContext(ctx).Model(&user).Where("mobile = ?", input.Identifier).Select(); err != nil {
				if err == pg.ErrNoRows {
					return nil, ErrInvalidCredentials
				}
				return nil, err
			}
		default:
			return nil, ErrLoginMethodNotFound
		}

		if !user.Enabled {
			return nil, ErrUserNotEnabled
		}

		cc.Logger().Infof("Identifier %s, method %s, login success", input.Identifier, input.Method)

		// generate token
		token, err := pie.AuthJwt(user.Id.String(), string(user.Role))
		if err != nil {
			return nil, err
		}
		return &LoginOutput{
			Token: token,
		}, nil
	}
}
