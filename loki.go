package zerologlokipublisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type LokiConfig struct {
	PushIntveralSeconds int
	// This will also trigger the send event
	MaxBatchSize int
	Values       map[string][][]string
	LokiEndpoint string
	BatchCount   int
	ServiceName  string
}

type lokiClient struct {
	config *LokiConfig
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
		if time.Now().Second()-lastRunTimestamp > l.config.PushIntveralSeconds || l.config.BatchCount > l.config.MaxBatchSize {
			// Loop over all log levels and send them
			for k, _ := range l.config.Values {
				if len(l.config.Values) > 0 {
					prevLogs := l.config.Values[k]
					l.config.Values[k] = [][]string{}
					err := pushToLoki(prevLogs, l.config.LokiEndpoint, k, l.config.ServiceName)
					if err != nil && isWorking {
						isWorking = false
						log.Error().Msgf("Logs are currently not being forwarded to loki due to an error: %v", err)
					}
					if err == nil && !isWorking {
						isWorking = true
						// I will not accept PR comments about this log message tyvm
						log.Info().Msgf("Logs publishing now functional again. Logs are being published to loki instance")
					}
				}
			}
			lastRunTimestamp = time.Now().Second()
			l.config.BatchCount = 0
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
