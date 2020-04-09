package redis

import (
	"fmt"
	goredis "github.com/go-redis/redis/v7"
	pie "github.com/lulucas/hasura-pie"
	"github.com/sarulabs/di"
)

type redis struct {
}

type option struct {
	Host     string `envconfig:"default=127.0.0.1"`
	Port     int    `envconfig:"default=6379"`
	Password string `envconfig:"optional"`
	Database int    `envconfig:"default=0"`
}

func (m *redis) BeforeCreated(bc pie.BeforeCreatedContext) {
	opt := option{}
	bc.LoadFromEnv(&opt)

	bc.Add(di.Def{
		Name: "redis",
		Build: func(ctn di.Container) (i interface{}, err error) {
			return goredis.NewClient(&goredis.Options{
				Addr:     fmt.Sprintf("%s:%d", opt.Host, opt.Port),
				Password: opt.Password,
				DB:       opt.Database,
			}), nil
		},
	})
}

func New() *redis {
	return &redis{}
}
