package finance

import (
	"github.com/lulucas/hasura-pie"
)

type finance struct {
}

func New() *finance {
	return &finance{}
}

func (m *finance) BeforeCreated(bc pie.BeforeCreatedContext) {

}

func (m *finance) Created(cc pie.CreatedContext) {

}
