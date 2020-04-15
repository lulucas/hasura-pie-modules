package account

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func createDefaultAdmin(cc pie.CreatedContext, opt option) error {
	if opt.DefaultAdminName != "" && opt.DefaultAdminPassword != "" {
		if err := validation.ValidateStruct(&opt,
			validation.Field(&opt.DefaultAdminName, validation.Required, validation.Length(5, 32), is.Alphanumeric),
			validation.Field(&opt.DefaultAdminPassword, validation.Required, validation.Length(8, 32), is.PrintableASCII),
		); err != nil {
			return err
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(opt.DefaultAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		dbCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		count, err := cc.DB().WithContext(dbCtx).Model(&model.User{}).Count()
		if err != nil {
			return err
		}
		if count == 0 {
			if err := cc.DB().WithContext(dbCtx).Insert(&model.User{
				Name:     opt.DefaultAdminName,
				Password: string(hashedPassword),
				Role:     model.RoleAdmin,
				Enabled:  true,
			}); err != nil {
				return err
			}
			cc.Logger().Warnf("Create default admin, name: %s", opt.DefaultAdminName)
		}
	}

	return nil
}
