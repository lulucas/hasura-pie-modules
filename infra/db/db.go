package db

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/lulucas/hasura-pie"
	"github.com/sarulabs/di"
)

type db struct {
}

type option struct {
	Host     string `envconfig:"default=127.0.0.1"`
	Port     int    `envconfig:"default=5432"`
	User     string `envconfig:"default=postgres"`
	Password string `envconfig:"optional"`
	Database string `envconfig:"default=postgres"`
}

func New() *db {
	return &db{}
}

func (m *db) BeforeCreated(bc pie.BeforeCreatedContext) {
	opt := option{}
	bc.LoadFromEnv(&opt)

	bc.Add(di.Def{
		Name: "db",
		Build: func(ctn di.Container) (i interface{}, err error) {
			return pg.Connect(&pg.Options{
				Addr:     fmt.Sprintf("%s:%d", opt.Host, opt.Port),
				User:     opt.User,
				Password: opt.Password,
				Database: opt.Database,
			}), nil
		},
	})
}

func (m *db) Created(cc pie.CreatedContext) {

}
