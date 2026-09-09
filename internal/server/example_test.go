package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"go.uber.org/zap"

	"github.com/shigabutdinoff/metrics/internal/storage"
)

// Полный цикл: приём метрик, чтение счётчика и список через роутер.
func Example() {
	s := New(storage.NewMemStorage(), zap.NewNop())
	s.setupRoutes()

	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	post := func(path, body string) string {
		resp, err := http.Post(ts.URL+path, "application/json", strings.NewReader(body))
		if err != nil {
			return err.Error()
		}
		defer resp.Body.Close()

		out, err := io.ReadAll(resp.Body)
		if err != nil {
			return err.Error()
		}
		return strings.TrimSpace(string(out))
	}

	fmt.Println(post("/update/", `{"id":"Alloc","type":"gauge","value":12.5}`))
	fmt.Println(post("/update/", `{"id":"PollCount","type":"counter","delta":5}`))
	fmt.Println(post("/update/", `{"id":"PollCount","type":"counter","delta":10}`))

	fmt.Println(post("/value/", `{"id":"PollCount","type":"counter"}`))

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	page, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(page))

	// Output:
	// {"id":"Alloc","type":"gauge","value":12.5}
	// {"id":"PollCount","type":"counter","delta":5}
	// {"id":"PollCount","type":"counter","delta":10}
	// {"id":"PollCount","type":"counter","delta":15}
	// <html><body><ul><li>counter PollCount: 15</li><li>gauge Alloc: 12.5</li></ul></body></html>
}
