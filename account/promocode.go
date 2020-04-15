package account

import (
	"context"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account/model"
	uuid "github.com/satori/go.uuid"
	"strings"
)

func initUserPromoCode(cc pie.CreatedContext) interface{} {
	type Event struct {
		pie.Event
		Old model.User
		New model.User
	}
	return func(ctx context.Context, evt Event) error {
		if _, err := cc.DB().Model(&evt.New).
			Where("id = ?", evt.New.Id).
			Where("promo_code = ?", nil).
			Set("promo_code = ?", strings.ReplaceAll(uuid.NewV4().String(), "-", "")[:11]).
			Update(); err != nil {
			return err
		}
		return nil
	}
}
