package account

import (
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	"github.com/pkg/errors"
)

var (
	ErrNameExists             = errors.New("account.name.exists")
	ErrMobileExists           = errors.New("account.mobile.exists")
	ErrLoginMethodNotFound    = errors.New("account.login-method-not-found")
	ErrRegisterMethodNotFound = errors.New("account.register-method-not-found")
	ErrRoleNotFound           = errors.New("account.role-not-found")
	ErrInvalidCredentials     = errors.New("account.invalid-credentials")
	ErrUserNotEnabled         = errors.New("account.user-not-enabled")
	ErrCaptchaInvalid         = errors.New("account.captcha-invalid")
)

type account struct {
	opt option
}

type option struct {
	DefaultAdminName     string `envconfig:"optional"`
	DefaultAdminPassword string `envconfig:"optional"`

	LoginMethods      []model.LoginMethod `envconfig:"default=name"`
	LoginImageCaptcha bool                `envconfig:"default=false"`

	RegisterImageCaptcha bool                   `envconfig:"default=false"`
	RegisterMethods      []model.RegisterMethod `envconfig:"default=name"`
	RegisterRoles        []model.Role           `envconfig:"default=user;merchant"`

	UpdatePasswordSmsCaptcha bool `envconfig:"default=false"`
}

func (m *account) BeforeCreated(bc pie.BeforeCreatedContext) {
	bc.LoadFromEnv(&m.opt)
}

func (m *account) Created(cc pie.CreatedContext) {
	if err := createDefaultAdmin(cc, m.opt); err != nil {
		cc.Logger().Fatalf("Create default admin error, %s", err.Error())
	}

	cc.HandleAction("login", login(cc, m.opt))
	cc.HandleAction("register", register(cc, m.opt))
	cc.HandleAction("update_password", updatePassword(cc, m.opt))
	cc.HandleAction("refresh_token", refreshToken(cc))

	cc.HandleEvent("init_user_promo_code", initUserPromoCode(cc))
}

func New() *account {
	return &account{}
}
