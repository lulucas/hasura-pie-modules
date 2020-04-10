package analysis

import (
	"github.com/go-redis/redis/v7"
	"github.com/lulucas/hasura-pie"
)

type analysis struct {
	r *redis.Client
}

func New() *analysis {
	return &analysis{}
}

func (m *analysis) BeforeCreated(bc pie.BeforeCreatedContext) {

}

func (m *analysis) Created(cc pie.CreatedContext) {
	m.r = cc.Get("redis").(*redis.Client)

	cc.Http().GET("/visit/hit.gif", hit(m))
}
