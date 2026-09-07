package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"go.uber.org/zap"
)

func computeHMAC(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func handlerEcho(body string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != "" {
			_, _ = io.WriteString(w, body)
		}
	}
}

func TestMiddleware_Request(t *testing.T) {
	const key = "secret"
	body := []byte(`{"id":"cpu","type":"gauge","value":1.5}`)
	correctHash := computeHMAC([]byte(key), body)

	tests := []struct {
		name       string
		key        string
		body       []byte
		hashHeader string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "пустой ключ - middleware прозрачен",
			key:        "",
			body:       body,
			hashHeader: "wrong-hash",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "верный хэш - запрос проходит",
			key:        key,
			body:       body,
			hashHeader: correctHash,
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "неверный хэш - 400",
			key:        key,
			body:       body,
			hashHeader: "deadbeef",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
		{
			name:       "нет заголовка хэша - запрос проходит",
			key:        key,
			body:       body,
			hashHeader: "",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "нет тела - запрос проходит",
			key:        key,
			body:       nil,
			hashHeader: "",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			h := Middleware(tt.key, zap.NewNop())(next)

			var reqBody io.Reader
			if tt.body != nil {
				reqBody = bytes.NewReader(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/updates/", reqBody)
			if tt.hashHeader != "" {
				req.Header.Set(header, tt.hashHeader)
			}
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидается %d", rr.Code, tt.wantStatus)
			}
			if called != tt.wantCalled {
				t.Fatalf("хендлер вызван = %v, ожидается %v", called, tt.wantCalled)
			}
		})
	}
}

type brokenBody struct{}

func (brokenBody) Read([]byte) (int, error) { return 0, errors.New("поток повреждён") }
func (brokenBody) Close() error             { return nil }

func TestMiddleware_BrokenBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("хендлер не должен вызываться")
	})
	h := Middleware("secret", zap.NewNop())(next)

	req := httptest.NewRequest(http.MethodPost, "/updates/", brokenBody{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("статус = %d, ожидается %d", rr.Code, http.StatusBadRequest)
	}
}

func TestMiddleware_Response(t *testing.T) {
	const key = "secret"
	responseBody := `{"id":"cpu","type":"gauge","value":1.5}`
	expectedHash := computeHMAC([]byte(key), []byte(responseBody))

	tests := []struct {
		name         string
		key          string
		responseBody string
		wantHash     string
	}{
		{
			name:         "ответ с телом - заголовок HashSHA256 выставлен",
			key:          key,
			responseBody: responseBody,
			wantHash:     expectedHash,
		},
		{
			name:         "ответ без тела - заголовок HashSHA256 выставлен для пустого тела",
			key:          key,
			responseBody: "",
			wantHash:     computeHMAC([]byte(key), []byte("")),
		},
		{
			name:         "пустой ключ - заголовок HashSHA256 отсутствует",
			key:          "",
			responseBody: responseBody,
			wantHash:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Middleware(tt.key, zap.NewNop())(handlerEcho(tt.responseBody, http.StatusOK))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			gotHash := rr.Header().Get(header)
			if gotHash != tt.wantHash {
				t.Fatalf("HashSHA256 = %q, ожидается %q", gotHash, tt.wantHash)
			}
		})
	}
}

func TestMiddleware_BodyAvailableAfterCheck(t *testing.T) {
	const key = "secret"
	body := []byte(`{"id":"cpu","type":"gauge","value":1.5}`)
	correctHash := computeHMAC([]byte(key), body)

	var receivedBody []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	h := Middleware(key, zap.NewNop())(next)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	req.Header.Set(header, correctHash)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if !bytes.Equal(receivedBody, body) {
		t.Fatalf("хендлер получил тело %q, ожидается %q", receivedBody, body)
	}
}

func echoBody(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, r.Body)
}

func TestMiddleware_Concurrent(t *testing.T) {
	const key = "secret"
	h := Middleware(key, zap.NewNop())(http.HandlerFunc(echoBody))

	var wg sync.WaitGroup
	for g := range 64 {
		wg.Go(func() {
			body := fmt.Appendf(nil, `{"id":"m%d","type":"gauge","value":%d}`, g, g)
			want := computeHMAC([]byte(key), body)

			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
			req.Header.Set(header, want)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("горутина %d: статус = %d, ожидается 200", g, rr.Code)
				return
			}
			if !bytes.Equal(rr.Body.Bytes(), body) {
				t.Errorf("горутина %d: тело ответа %q, ожидается %q", g, rr.Body.Bytes(), body)
			}
			if got := rr.Header().Get(header); got != want {
				t.Errorf("горутина %d: HashSHA256 = %q, ожидается %q", g, got, want)
			}
		})
	}
	wg.Wait()
}

func BenchmarkMiddleware(b *testing.B) {
	body := bytes.Repeat([]byte(`{"id":"Alloc","type":"gauge","value":2457600},`), 40)
	hash := computeHMAC([]byte("secret"), body)

	run := func(b *testing.B, key string) {
		h := Middleware(key, zap.NewNop())(http.HandlerFunc(echoBody))
		rd := bytes.NewReader(body)
		req := httptest.NewRequest(http.MethodPost, "/updates/", rd)
		req.Header.Set(header, hash)
		reqBody := req.Body

		b.ReportAllocs()
		for b.Loop() {
			rd.Reset(body)
			req.Body = reqBody
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				b.Fatalf("статус = %d, ожидается %d", rr.Code, http.StatusOK)
			}
		}
	}

	b.Run("с ключом", func(b *testing.B) { run(b, "secret") })
	b.Run("без ключа", func(b *testing.B) { run(b, "") })
}
