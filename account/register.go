package account

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Identifier       string
	Password         string
	Method           model.RegisterMethod
	Role             model.Role
	ImageCaptchaId   string
	ImageCaptchaCode string
	SmsCaptchaCode   *string
	PromoCode        *string
}

type RegisterOutput struct {
	Token        string
	RefreshToken string
}

func register(cc pie.CreatedContext, opt option) interface{} {
	c := cc.Get("captcha").(Captcha)

	return func(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
		if err := validation.ValidateStruct(&input,
			validation.Field(&input.Identifier, validation.Required, validation.Length(5, 32)),
			validation.Field(&input.Password, validation.Required, validation.Length(6, 32)),
		); err != nil {
			return nil, err
		}

		// check role validation
		if !input.Role.In(opt.RegisterRoles...) {
			return nil, ErrRoleNotFound
		}

		// check register method validation
		if !input.Method.In(opt.RegisterMethods...) {
			return nil, ErrRegisterMethodNotFound
		}

		if opt.RegisterImageCaptcha {
			if err := c.ValidateImageCaptcha(input.ImageCaptchaId, input.ImageCaptchaCode); err != nil {
				return nil, err
			}
		}

		user := model.User{}

		tx, err := cc.DB().WithContext(ctx).Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		switch input.Method {
		case model.RegisterMethodMobile:
			// sms captcha
			if input.SmsCaptchaCode == nil {
				return nil, ErrCaptchaInvalid
			}
			if err := c.ValidateSmsCaptcha(input.Identifier, *input.SmsCaptchaCode); err != nil {
				return nil, err
			}

			// find user
			if err := tx.Model(&user).Limit(1).Where("mobile = ?", input.Identifier).Select(); err == nil {
				return nil, ErrMobileExists
			}

			user.Name = "m" + input.Identifier
		case model.RegisterMethodName:
			user.Name = input.Identifier
			// find user
			if err := tx.Model(&user).Limit(1).Where("name = ?", input.Identifier).Select(); err == nil {
				return nil, ErrNameExists
			}
		default:
			return nil, ErrRegisterMethodNotFound
		}

		// promo
		parent := model.User{}
		if input.PromoCode != nil {
			if err := tx.Model(&parent).Where("promo_code = ?", input.PromoCode).Select(); err != nil {
				if err == pg.ErrNoRows {
					// ignore
				} else {
					return nil, err
				}
			} else {
				// set parent
				user.ParentId = &parent.Id
			}
		}

		// set user
		password, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Mobile = &input.Identifier
		user.Password = string(password)
		user.Role = input.Role
		user.Enabled = true
		// insert user
		if err := tx.Insert(&user); err != nil {
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		// generate token
		token, err := pie.AuthJwt(user.Id.String(), string(user.Role), TokenDuration)
		if err != nil {
			return nil, err
		}
		refreshToken, err := pie.AuthJwt(user.Id.String(), string(user.Role), RefreshTokenDuration)
		if err != nil {
			return nil, err
		}
		return &RegisterOutput{
			Token:        token,
			RefreshToken: refreshToken,
		}, nil
	}
}
