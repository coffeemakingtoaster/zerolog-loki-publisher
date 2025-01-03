package zerologlokipublisher

import (
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// Compile time check if lokiHook (still) satisfies the Hook interface from zerolog
var _ zerolog.Hook = (*lokiHook)(nil)

func NewHook(config LokiConfig) *lokiHook {
	client := lokiClient{config: &config, done: make(chan bool)}
	go client.bgRun()
	return &lokiHook{client: &client}
}

type lokiHook struct {
	client *lokiClient
}

func (h *lokiHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	h.client.config.Values[level.String()] = append(h.client.config.Values[level.String()], []string{strconv.FormatInt(time.Now().UnixNano(), 10), msg})
	h.client.config.BatchCount++
}

func (h *lokiHook) Stop() {
	// gracefull shutdown
	h.client.done <- true
}
