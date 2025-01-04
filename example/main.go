package main

import (
	"time"

	zerologlokipublisher "github.com/coffeemakingtoaster/zerolog-loki-publisher"
	"github.com/rs/zerolog/log"
)

func main() {
	hook := zerologlokipublisher.NewHook(zerologlokipublisher.LokiConfig{
		PushIntveralSeconds: 10,
		MaxBatchSize:        500,
		LokiEndpoint:        "http://localhost:3100",
		BatchCount:          0,
	})

	log.Logger = log.Hook(hook)

	for {
		log.Info().Msg("Sample log message")
		time.Sleep(1 * time.Second)
	}
}
