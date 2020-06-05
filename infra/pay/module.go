package pay

import (
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel"
	"github.com/sarulabs/di"
	"reflect"
)

type pay struct {
	db       *pg.DB
	channels map[string]channel.Channel
}

func New(channels ...channel.Channel) *pay {
	chs := map[string]channel.Channel{}
	for _, ch := range channels {
		name := reflect.TypeOf(ch).Elem().Name()
		chs[name] = ch
	}
	return &pay{
		channels: chs,
	}
}

func (m *pay) BeforeCreated(bc pie.BeforeCreatedContext) {
	bc.Add(di.Def{
		Name: "pay",
		Build: func(ctn di.Container) (i interface{}, err error) {
			return m, nil
		},
	})
}

func (m *pay) Created(cc pie.CreatedContext) {
	m.db = cc.DB()

	cc.Rest().POST("/notify/:section/:id", notify(cc, m.channels))
	cc.Rest().POST("/notify/:section/:id/:uid", notify(cc, m.channels))
}
