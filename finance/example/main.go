package main

import (
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/finance"
)

func main() {
	app := pie.NewApp()
	app.AddModule(
		finance.New(),
	)
	app.Start()
}
