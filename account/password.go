package account

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	"golang.org/x/crypto/bcrypt"
)

func updatePassword(cc pie.CreatedContext, opt option) interface{} {
	c := cc.Get("captcha").(Captcha)

	type UpdatePasswordOutput struct {
		Token string
	}

	return func(ctx context.Context, input struct {
		Password   string
		SmsCaptcha *string
	}) (*UpdatePasswordOutput, error) {
		if err := validation.ValidateStruct(&input,
			validation.Field(&input.Password, validation.Required, validation.Length(6, 32)),
		); err != nil {
			return nil, err
		}

		user := model.User{}
		userId := cc.GetSession(ctx).UserId
		if err := cc.DB().WithContext(ctx).Model(&user).Where("id = ?", userId).Select(); err != nil {
			if err == pg.ErrNoRows {
				return nil, ErrInvalidCredentials
			}
			return nil, err
		}

		// sms captcha
		if opt.UpdatePasswordSmsCaptcha {
			if user.Mobile == nil || input.SmsCaptcha == nil {
				return nil, ErrCaptchaInvalid
			}
			if err := c.ValidateSmsCaptcha(*user.Mobile, *input.SmsCaptcha); err != nil {
				return nil, err
			}
		}

		password, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		user.Password = string(password)

		if _, err := cc.DB().WithContext(ctx).
			Model(&user).
			Where("id = ?", userId).
			Set("password = ?", user.Password).Update(); err != nil {
			return nil, err
		}

		cc.Logger().Infof("User id: %s, name: %s, change password success", user.Id, user.Name)

		// generate token
		token, err := pie.AuthJwt(user.Id.String(), string(user.Role))
		if err != nil {
			return nil, err
		}
		return &UpdatePasswordOutput{
			Token: token,
		}, nil
	}
}

// TODO RECOVER PASSWORD