package pay

import (
	"github.com/go-pg/pg/v9"
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel"
	"github.com/lulucas/hasura-pie-modules/infra/pay/channel/bf"
	"github.com/sarulabs/di"
	"github.com/shopspring/decimal"
)

type pay struct {
	pie.DefaultModule
	db       *pg.DB
	channels map[string]channel.Channel
}

func New() *pay {
	// 初始化支持的通道
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
	m.db = cc.Get("postgres").(*pg.DB)

	// 回调
	cc.Http().POST("/pay/notify/:id/:uid", notify(cc, m.channels))
}

func (m *pay) Pay(id int, amount decimal.Decimal, returnUrl string) error {

	return nil
}
