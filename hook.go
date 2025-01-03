package zerologlokipublisher

import (
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func NewHook(config *LokiConfig) *lokiHook {
	client := lokiClient{config: config}
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
