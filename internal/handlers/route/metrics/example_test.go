package metrics_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/shigabutdinoff/metrics/internal/handlers/route/metrics"
	"github.com/shigabutdinoff/metrics/internal/storage"
)

// Список всех метрик. Строки отсортированы, поэтому счётчики идут первыми.
func ExampleIndex() {
	st := storage.NewMemStorage()
	ctx := context.Background()

	alloc := 12.5
	st.SetGauge(ctx, "Alloc", &alloc)
	delta := int64(15)
	st.AddCounter(ctx, "PollCount", &delta)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	metrics.Index(st).ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Content-Type"))
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// text/html; charset=utf-8
	// <html><body><ul><li>counter PollCount: 15</li><li>gauge Alloc: 12.5</li></ul></body></html>
}
