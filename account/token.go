package account

import (
	"context"
	pie "github.com/lulucas/hasura-pie"
)

func refreshToken(cc pie.CreatedContext) interface{} {
	type RefreshTokenOutput struct {
		Token string
	}
	return func(ctx context.Context) (*RefreshTokenOutput, error) {
		session := cc.GetSession(ctx)
		userId := session.UserId
		if userId == nil {
			return nil, ErrInvalidCredentials
		}
		// refresh token
		token, err := pie.AuthJwt(userId.String(), session.Role, TokenDuration)
		if err != nil {
			return nil, err
		}
		return &RefreshTokenOutput{
			Token: token,
		}, nil
	}
}
