package healthcheck_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/shigabutdinoff/metrics/internal/handlers/route/healthcheck"
)

// Проверка соединения с БД: без DATABASE_DSN хендлер отдаёт 500.
func ExamplePing() {
	h := healthcheck.Ping(func() *sql.DB { return nil })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 500
	// Не удалось открыть соединение с БД
}
