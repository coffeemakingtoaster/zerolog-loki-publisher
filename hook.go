package zerologlokipublisher

import (
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Compile time check if lokiHook (still) satisfies the Hook interface from zerolog
var _ zerolog.Hook = (*lokiHook)(nil)

// Instantiate a new instance of the hook.
// This includes starting the background go routine for publishing log messages to loki
// Returns a pointer to the hook that will have to be passed to zerolog
func NewHook(config LokiConfig) *lokiHook {
	client := lokiClient{config: &config, done: make(chan bool), values: sync.Map{}, batchCount: 0}
	go client.bgRun()
	return &lokiHook{client: &client}
}

type lokiHook struct {
	client *lokiClient
}

func (h *lokiHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	curr_val, err := h.client.values.Load(level.String())
	if !err || curr_val == nil {
		curr_val = [][]string{}
	}
	h.client.values.Store(level.String(), append(curr_val.([][]string), []string{strconv.FormatInt(time.Now().UnixNano(), 10), msg}))
	h.client.batchCount++
}

func (h *lokiHook) Stop() {
	// gracefull shutdown
	h.client.done <- true
}
