package main

import (
	pie "github.com/lulucas/hasura-pie"
	"github.com/lulucas/hasura-pie-modules/analysis"
	"github.com/lulucas/hasura-pie-modules/infra/redis"
)

func main() {
	app := pie.NewApp()
	app.AddModule(
		redis.New(),
		analysis.New(),
	)
	app.Start()
}
