package zerologlokipublisher

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func startMockLoki() (*sync.WaitGroup, *http.Server) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/loki/api/v1/push",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("got data")
			w.WriteHeader(http.StatusOK)
		})
	srv := http.Server{Addr: "127.0.0.1:3100", Handler: serveMux}
	go func() {
		defer wg.Done()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("ListenAndServe(): %v", err))
		}
	}()

	return wg, &srv
}

func publishToClient(client lokiClient) {

	msgs := [][]string{
		{strconv.FormatInt(time.Now().UnixNano(), 10), "Sample1"},
		{strconv.FormatInt(time.Now().UnixNano(), 10), "Sample2"},
		{strconv.FormatInt(time.Now().UnixNano(), 10), "Sample3"},
	}

	client.config.Values.Store("Debug", msgs)
	client.config.BatchCount += 3
}

func testClientClear(client lokiClient, t *testing.T) {
	startTime := time.Now().Second()

	for time.Now().Second()-startTime < 10 {
		curr, ok := client.config.Values.Load("Debug")
		if !ok {
			return
		}
		if curr == nil {
			return
		}
		if len(curr.([][]string)) == 0 {
			return
		}
	}

	t.Error("Publish to loki did not run properly")
}

func TestMain(m *testing.M) {
	wg, srv := startMockLoki()

	m.Run()

	srv.Close()
	wg.Wait()
}

func Test_triggerSendViaTime(t *testing.T) {

	config := LokiConfig{
		PushIntveralSeconds: 1,
		MaxBatchSize:        50000,
		LokiEndpoint:        "http://127.0.0.1:3100",
		BatchCount:          0,
		Values:              sync.Map{},
	}

	client := lokiClient{config: &config, done: make(chan bool)}

	go client.bgRun()

	publishToClient(client)

	testClientClear(client, t)
}

func Test_triggerSendViaItemCount(t *testing.T) {
	config := LokiConfig{
		PushIntveralSeconds: 50000,
		MaxBatchSize:        3,
		LokiEndpoint:        "http://127.0.0.1:3100",
		BatchCount:          0,
		Values:              sync.Map{},
	}
	client := lokiClient{config: &config, done: make(chan bool)}

	go client.bgRun()

	publishToClient(client)

	testClientClear(client, t)
}
