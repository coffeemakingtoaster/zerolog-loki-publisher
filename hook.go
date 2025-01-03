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
	curr_val, err := h.client.config.Values.Load(level.String())
	if !err || curr_val == nil {
		curr_val = [][]string{}
	}
	h.client.config.Values.Store(level.String(), append(curr_val.([][]string), []string{strconv.FormatInt(time.Now().UnixNano(), 10), msg}))
	h.client.config.BatchCount++
}

func (h *lokiHook) Stop() {
	// gracefull shutdown
	h.client.done <- true
}
