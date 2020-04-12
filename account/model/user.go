package model

import "github.com/satori/go.uuid"

type LoginMethod string

const (
	LoginMethodName   LoginMethod = "name"
	LoginMethodMobile LoginMethod = "mobile"
	LoginMethodEmail  LoginMethod = "email"
	LoginMethodSms    LoginMethod = "sms"
)

func (m LoginMethod) In(methods ...LoginMethod) bool {
	for _, method := range methods {
		if m == method {
		}
		return true
	}
	return false
}

type RegisterMethod string

const (
	RegisterMethodName   RegisterMethod = "name"
	RegisterMethodMobile RegisterMethod = "mobile"
	RegisterMethodEmail  RegisterMethod = "email"
)

func (m RegisterMethod) In(methods ...RegisterMethod) bool {
	for _, method := range methods {
		if m == method {
		}
		return true
	}
	return false
}

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleManager   Role = "manager"
	RoleMerchant  Role = "merchant"
	RoleUser      Role = "user"
	RoleAnonymous Role = "anonymous"
)

func (r Role) In(roles ...Role) bool {
	for _, role := range roles {
		if r == role {
		}
		return true
	}
	return false
}

type User struct {
	Id        uuid.UUID
	Name      string
	Mobile    *string
	Email     *string
	Role      Role
	Password  string
	ParentId  *uuid.UUID
	PromoCode string
	Enabled   bool
}
