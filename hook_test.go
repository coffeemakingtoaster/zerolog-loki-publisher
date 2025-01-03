package zerologlokipublisher_test

import (
	"runtime"
	"testing"
	"time"

	zerologlokipublisher "github.com/coffeemakingtoaster/zerolog-loki-publisher"
)

func Test_bgJobControl(t *testing.T) {
	initialGoRountineCount := runtime.NumGoroutine()
	hook := zerologlokipublisher.NewHook(zerologlokipublisher.LokiConfig{
		PushIntveralSeconds: 10,  // Threshhold of 10s
		MaxBatchSize:        500, //Threshold of 500 events
		LokiEndpoint:        "127.0.0.0",
		BatchCount:          0,
		Values:              make(map[string][][]string),
	})

	if runtime.NumGoroutine() != initialGoRountineCount+1 {
		t.Errorf("Expected bg job to have been started. (Wanted: %d goroutines, Got: %d goroutines)", initialGoRountineCount+1, runtime.NumGoroutine())
	}

	hook.Stop()

	startTime := time.Now().Second()

	for time.Now().Second()-startTime < 10 {
		// Success
		if runtime.NumGoroutine() == initialGoRountineCount {
			return
		}
	}

	t.Errorf("Expected bg job to have been stopped. (Wanted: %d goroutines, Got: %d goroutines)", initialGoRountineCount+1, runtime.NumGoroutine())
}
