package ip2region

import (
	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
	pie "github.com/lulucas/hasura-pie"
	"github.com/sarulabs/di"
)

type ip struct {
}

func New() *ip {
	return &ip{}
}

func (m *ip) BeforeCreated(bc pie.BeforeCreatedContext) {
	bc.Add(di.Def{
		Name: "ip2region",
		Build: func(ctn di.Container) (interface{}, error) {
			return ip2region.New("ip2region.db")
		},
		Close: func(obj interface{}) error {
			obj.(*ip2region.Ip2Region).Close()
			return nil
		},
	})
}

func (m *ip) Created(cc pie.CreatedContext) {

}
