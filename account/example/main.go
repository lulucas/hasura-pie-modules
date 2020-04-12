package main

import (
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/account"
)

func main() {
	app := pie.NewApp()
	app.AddModule(
		account.New(),
	)
	app.Start()
}
