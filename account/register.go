package account

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type RegisterInput struct {
	Identifier       string
	Password         string
	Method           model.RegisterMethod
	ImageCaptchaId   string
	ImageCaptchaCode string
	SmsCaptchaCode   *string
	PromoCode        *string
}

type RegisterOutput struct {
	Token string
}

func register(cc pie.CreatedContext, opt option) interface{} {
	db := cc.Get("postgres").(*pg.DB)
	c := cc.Get("captcha").(Captcha)

	return func(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
		if err := validation.ValidateStruct(&input,
			validation.Field(&input.Identifier, validation.Required, validation.Length(5, 32)),
			validation.Field(&input.Password, validation.Required, validation.Length(6, 32)),
		); err != nil {
			return nil, err
		}

		// 图形验证码
		if opt.RegisterImageCaptcha {
			if err := c.ValidateImageCaptcha(input.ImageCaptchaId, input.ImageCaptchaCode); err != nil {
				return nil, err
			}
		}

		user := model.User{}

		tx, err := db.WithContext(ctx).Begin()
		if err != nil {
			return nil, err
		}

		switch input.Method {
		case model.RegisterMethodMobile:
			// 短信验证码
			if input.SmsCaptchaCode == nil {
				return nil, ErrCaptchaInvalid
			}
			if err := c.ValidateSmsCaptcha(input.Identifier, *input.SmsCaptchaCode); err != nil {
				return nil, err
			}

			// 查询用户是否已存在
			if err := tx.Model(&user).Limit(1).Where(string(input.Method)+" = ?", input.Identifier).Select(); err == nil {
				return nil, ErrMobileExists
			}

			// 如果有邀请码则设置上级用户
			parent := model.User{}
			if input.PromoCode != nil {
				if err := tx.Model(&parent).Where("promo_code = ?", input.PromoCode).Select(); err != nil {
					if err == pg.ErrNoRows {
						// 如果邀请码无效则忽略
					} else {
						return nil, err
					}
				} else {
					// 设置上级
					user.ParentId = &parent.Id
				}
			}

			// 设置用户参数
			password, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, err
			}
			user.Name = "m" + input.Identifier
			user.Mobile = &input.Identifier
			user.Password = string(password)
			user.Role = model.RoleUser
			user.PromoCode = strings.ReplaceAll(uuid.NewV4().String(), "-", "")[:11]
			user.Enabled = true
			// 插入用户
			if err := tx.Insert(&user); err != nil {
				return nil, err
			}
		default:
			return nil, ErrRegisterMethodNotFound
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		token, err := pie.AuthJwt(user.Id.String(), string(user.Role))
		if err != nil {
			return nil, err
		}
		return &RegisterOutput{
			Token: token,
		}, nil
	}
}
