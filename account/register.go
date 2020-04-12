package account

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	"github.com/lulucas/hasura-pie-modules/account/utils"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
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
	Token string
}

func register(cc pie.CreatedContext, opt option) interface{} {
	c := cc.Get("captcha").(Captcha)

	return func(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
		if err := validation.ValidateStruct(&input,
			validation.Field(&input.Identifier, validation.Required, validation.Length(5, 32)),
			validation.Field(&input.Password, validation.Required, validation.Length(6, 32)),
			validation.Field(&input.Role, validation.Required, validation.In(utils.StringSlice2InterfaceSlice(opt.RegisterRoles)...)),
		); err != nil {
			return nil, err
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
			if err := tx.Model(&user).Limit(1).Where(string(input.Method)+" = ?", input.Identifier).Select(); err == nil {
				return nil, ErrMobileExists
			}

			// register by promo code
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
			user.Name = "m" + input.Identifier
			user.Mobile = &input.Identifier
			user.Password = string(password)
			user.Role = input.Role
			user.PromoCode = strings.ReplaceAll(uuid.NewV4().String(), "-", "")[:11]
			user.Enabled = true
			// insert user
			if err := tx.Insert(&user); err != nil {
				return nil, err
			}
		default:
			return nil, ErrRegisterMethodNotFound
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		// generate token
		token, err := pie.AuthJwt(user.Id.String(), string(user.Role))
		if err != nil {
			return nil, err
		}
		return &RegisterOutput{
			Token: token,
		}, nil
	}
}
