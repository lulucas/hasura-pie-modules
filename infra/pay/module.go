package pay

import (
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel/bf"
	"github.com/sarulabs/di"
)

type pay struct {
	db       *pg.DB
	channels map[string]channel.Channel
}

func New() *pay {
	channels := map[string]channel.Channel{
		"bf": bf.New(),
	}
	return &pay{
		channels: channels,
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

	cc.Rest().POST("/notify/:id", notify(cc, m.channels))
	cc.Rest().POST("/notify/:id/:uid", notify(cc, m.channels))
}
