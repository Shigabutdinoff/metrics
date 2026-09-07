package update_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/shigabutdinoff/metrics/internal/handlers/route/update"
	"github.com/shigabutdinoff/metrics/internal/storage"
)

// Приём метрики из пути: параметры читает chi, поэтому нужен роутер.
func ExampleStoreTextPlain() {
	st := storage.NewMemStorage()

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", update.StoreTextPlain(st))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12.5", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// OK
}

// Приём метрики в JSON: в ответе присланный объект, а не сумма.
func ExampleStoreApplicationJSON() {
	st := storage.NewMemStorage()
	h := update.StoreApplicationJSON(st)

	body := strings.NewReader(`{"id":"PollCount","type":"counter","delta":5}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Content-Type"))
	fmt.Print(rr.Body.String())

	// Output:
	// 200
	// application/json
	// {"id":"PollCount","type":"counter","delta":5}
}
