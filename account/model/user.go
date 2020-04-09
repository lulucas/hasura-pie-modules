package model

import "github.com/satori/go.uuid"

type LoginMethod string

const (
	LoginMethodName   LoginMethod = "name"
	LoginMethodMobile LoginMethod = "mobile"
	LoginMethodEmail  LoginMethod = "email"
	LoginMethodSms    LoginMethod = "sms"
)

type RegisterMethod string

const (
	RegisterMethodName   RegisterMethod = "name"
	RegisterMethodMobile RegisterMethod = "mobile"
	RegisterMethodEmail  RegisterMethod = "email"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleManager   Role = "manager"
	RoleMerchant  Role = "merchant"
	RoleUser      Role = "user"
	RoleAnonymous Role = "anonymous"
)

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
