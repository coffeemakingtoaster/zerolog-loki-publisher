package zerologlokipublisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Config for the loki pusblishing
type LokiConfig struct {
	PushIntveralSeconds int    // Threshhold in seconds. Once this time is exceeded all recoreded logs are sent to loki
	MaxBatchSize        int    // Threshhold in log amount. Once this threshold is reached all recoreded logs are sent to loki
	LokiEndpoint        string // Loki instance endpoint. This does not include the path
	ServiceName         string // Service name as forwarded to loki
}

type lokiClient struct {
	config *LokiConfig
	done   chan bool
	values sync.Map
	// Current amount of logs that have not yet been sent
	batchCount int
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type lokiLogEvent struct {
	Streams []lokiStream
}

func (l *lokiClient) bgRun() {
	lastRunTimestamp := 0
	isWorking := true
	for {
		if time.Now().Second()-lastRunTimestamp > l.config.PushIntveralSeconds || l.batchCount >= l.config.MaxBatchSize {
			// Loop over all log levels and send them
			l.values.Range(func(key, value any) bool {
				logMessages := value.([][]string)
				if len(logMessages) > 0 {
					prevLogs := logMessages
					l.values.Delete(key)
					err := pushToLoki(prevLogs, l.config.LokiEndpoint, key.(string), l.config.ServiceName)
					if err != nil && isWorking {
						isWorking = false
						log.Error().Msgf("Logs are currently not being forwarded to loki due to an error: %v", err)
					}
					if err == nil && !isWorking {
						isWorking = true
						log.Info().Msgf("Logs publishing now functional again. Logs are being published to loki instance")
					}
				}
				return true
			})
			lastRunTimestamp = time.Now().Second()
			l.batchCount = 0
		}
		if <-l.done {
			break
		}
	}
}

/*
This function contains *no* error handling/logging because this:
a) should not crash the application
b) would mean that every run of this creates further logs that cannot be published
=> The error will be returned and the problem will be logged ONCE by the handling function
*/
func pushToLoki(logs [][]string, lokiEndpoint, logLevel, serviceName string) error {
	lokiPushPath := "/loki/api/v1/push"

	data, err := json.Marshal(lokiLogEvent{
		Streams: []lokiStream{
			{
				Stream: map[string]string{
					"service": serviceName,
					"level":   logLevel,
				},
				Values: logs,
			},
		},
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", lokiEndpoint, lokiPushPath), bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)

	defer cancel()

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
