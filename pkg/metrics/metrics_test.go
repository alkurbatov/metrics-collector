//nolint:gocritic
package metrics_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

// This is an example of pushing counter metric to metrics collector.
func ExampleNewUpdateCounterReq() {
	// Record a counter metric.
	value := metrics.Counter(10)

	// Create new update counter request.
	data := metrics.NewUpdateCounterReq("ExampleCounter", value)

	// This is a HTTP handler for testing purposes.
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	// This is a HTTP server for testing purposes.
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Create JSON payload.
	payload, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	body := bytes.NewReader(payload)

	// Send an HTTP request to the test server.
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/update", body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// For this example we don't care about the response.
	defer resp.Body.Close()
}
