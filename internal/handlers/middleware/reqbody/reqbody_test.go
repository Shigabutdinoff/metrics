package reqbody

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type brokenBody struct{}

func (brokenBody) Read([]byte) (int, error) { return 0, errors.New("поток повреждён") }
func (brokenBody) Close() error             { return nil }

func TestRead(t *testing.T) {
	t.Run("превышение лимита даёт 413", func(t *testing.T) {
		body := strings.Repeat("a", MaxBodySize+1)
		req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(body))
		rr := httptest.NewRecorder()
		if _, ok := Read(rr, req); ok || rr.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("ok = %v, статус = %d; ожидается false, %d", ok, rr.Code, http.StatusRequestEntityTooLarge)
		}
	})

	t.Run("повреждённое тело даёт 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates/", brokenBody{})
		rr := httptest.NewRecorder()
		if _, ok := Read(rr, req); ok || rr.Code != http.StatusBadRequest {
			t.Fatalf("ok = %v, статус = %d; ожидается false, %d", ok, rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("тело остаётся доступным", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader("abc"))
		rr := httptest.NewRecorder()
		body, ok := Read(rr, req)
		if !ok || string(body) != "abc" {
			t.Fatalf("ok = %v, body = %q; ожидается true, %q", ok, body, "abc")
		}
		rest, _ := io.ReadAll(req.Body)
		if string(rest) != "abc" {
			t.Fatalf("r.Body = %q, ожидается %q", rest, "abc")
		}
	})

	t.Run("повторный Read отдаёт то же тело", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader("abc"))
		rr := httptest.NewRecorder()
		first, _ := Read(rr, req)
		second, ok := Read(rr, req)
		if !ok || &first[0] != &second[0] {
			t.Fatalf("ok = %v, второй Read вернул другой буфер", ok)
		}
		rest, _ := io.ReadAll(req.Body)
		if string(rest) != "abc" {
			t.Fatalf("r.Body после повторного Read = %q, ожидается %q", rest, "abc")
		}
	})
}

func TestError(t *testing.T) {
	t.Run("превышение лимита даёт 413", func(t *testing.T) {
		rr := httptest.NewRecorder()
		Error(rr, &http.MaxBytesError{Limit: 8})
		if rr.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("статус = %d, ожидается %d", rr.Code, http.StatusRequestEntityTooLarge)
		}
	})

	t.Run("прочая ошибка даёт 400 с текстом", func(t *testing.T) {
		rr := httptest.NewRecorder()
		Error(rr, errors.New("битый json"))
		if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "битый json") {
			t.Fatalf("статус = %d, тело = %q; ожидается 400 с текстом ошибки", rr.Code, rr.Body.String())
		}
	})
}
