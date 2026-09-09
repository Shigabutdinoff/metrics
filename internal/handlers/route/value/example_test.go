package value_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/shigabutdinoff/metrics/internal/handlers/route/value"
	"github.com/shigabutdinoff/metrics/internal/storage"
)

// Чтение метрики по параметрам пути. Ответ содержит только число.
func ExampleShowTextPlain() {
	st := storage.NewMemStorage()
	alloc := 12.5
	st.SetGauge(context.Background(), "Alloc", &alloc)

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", value.ShowTextPlain(st))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// 12.5
}

// Чтение метрики в JSON. Для счётчика возвращается накопленная сумма дельт.
func ExampleShowApplicationJSON() {
	st := storage.NewMemStorage()
	ctx := context.Background()
	first, second := int64(5), int64(10)
	st.AddCounter(ctx, "PollCount", &first)
	st.AddCounter(ctx, "PollCount", &second)

	h := value.ShowApplicationJSON(st)

	body := strings.NewReader(`{"id":"PollCount","type":"counter"}`)
	req := httptest.NewRequest(http.MethodPost, "/value/", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Print(rr.Body.String())

	// Output:
	// 200
	// {"id":"PollCount","type":"counter","delta":15}
}
