package compress

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func gzipBytes(p []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(p); err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func gunzip(p []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(p))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

func handlerEcho(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func benchBody() []byte {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := range 30 {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `{"id":"Gauge%02d","type":"gauge","value":%d.5}`, i, i*1000)
	}
	sb.WriteString(`,{"id":"PollCount","type":"counter","delta":42}]`)
	return []byte(sb.String())
}

func TestGzipMiddleware(t *testing.T) {
	body := []byte(`{"id":"cpu","type":"gauge","value":1.5}`)
	gz := gzipBytes(body)

	tests := []struct {
		name            string
		reqBody         []byte
		contentEncoding string
		acceptEncoding  string
		wantStatus      int
		wantGzipResp    bool
	}{
		{name: "без сжатия", reqBody: body, wantStatus: http.StatusOK},
		{name: "gzip-запрос", reqBody: gz, contentEncoding: "gzip", wantStatus: http.StatusOK},
		{name: "gzip-ответ", reqBody: body, acceptEncoding: "gzip", wantStatus: http.StatusOK, wantGzipResp: true},
		{name: "gzip в обе стороны", reqBody: gz, contentEncoding: "gzip", acceptEncoding: "gzip", wantStatus: http.StatusOK, wantGzipResp: true},
		{name: "битый gzip в теле", reqBody: []byte("not gzip"), contentEncoding: "gzip", wantStatus: http.StatusBadRequest},
		{name: "битый gzip в теле, gzip-ответ", reqBody: []byte("not gzip"), contentEncoding: "gzip", acceptEncoding: "gzip", wantStatus: http.StatusBadRequest},
	}

	h := GzipMiddleware()(http.HandlerFunc(handlerEcho))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(tt.reqBody))
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидается %d", rr.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				if rr.Body.Len() != 0 {
					t.Fatalf("тело ошибки = %q, ожидается пустое", rr.Body.Bytes())
				}
				if ce := rr.Header().Get("Content-Encoding"); ce != "" {
					t.Fatalf("Content-Encoding = %q, ожидается пустой", ce)
				}
				return
			}

			got := rr.Body.Bytes()
			if tt.wantGzipResp {
				if ce := rr.Header().Get("Content-Encoding"); ce != "gzip" {
					t.Fatalf("Content-Encoding = %q, ожидается gzip", ce)
				}
				var err error
				if got, err = gunzip(got); err != nil {
					t.Fatalf("gunzip() ошибка = %v", err)
				}
			}
			if !bytes.Equal(got, body) {
				t.Fatalf("тело = %q, ожидается %q", got, body)
			}
		})
	}
}

func roundTrip(h http.Handler, body []byte) error {
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(gzipBytes(body)))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return fmt.Errorf("статус = %d, ожидается %d", rr.Code, http.StatusOK)
	}
	got, err := gunzip(rr.Body.Bytes())
	if err != nil {
		return fmt.Errorf("gunzip() ошибка = %w", err)
	}
	if !bytes.Equal(got, body) {
		return errors.New("тело ответа не совпадает с телом запроса")
	}
	return nil
}

func TestGzipMiddleware_Reuse(t *testing.T) {
	h := GzipMiddleware()(http.HandlerFunc(handlerEcho))

	bodies := make([][]byte, 20)
	for i := range bodies {
		bodies[i] = bytes.Repeat([]byte(fmt.Sprintf(`{"n":%d}`, i)), i*40+1)
	}

	// после ошибки Reset объекты пула должны остаться пригодными
	req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("битый gzip: статус = %d, ожидается 400", rr.Code)
	}

	for i, body := range bodies {
		if err := roundTrip(h, body); err != nil {
			t.Fatalf("запрос %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	for g := range 32 {
		wg.Go(func() {
			for j := range 10 {
				if err := roundTrip(h, bodies[(g+j)%len(bodies)]); err != nil {
					t.Errorf("горутина %d, запрос %d: %v", g, j, err)
					return
				}
			}
		})
	}
	wg.Wait()
}

func BenchmarkGzipMiddleware_Response(b *testing.B) {
	body := benchBody()
	h := GzipMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ReportAllocs()
	for b.Loop() {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("статус = %d, ожидается %d", rr.Code, http.StatusOK)
		}
	}
}

func BenchmarkGzipMiddleware_Request(b *testing.B) {
	gz := gzipBytes(benchBody())
	h := GzipMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	rd := bytes.NewReader(gz)
	req := httptest.NewRequest(http.MethodPost, "/updates/", rd)
	req.Header.Set("Content-Encoding", "gzip")

	b.ReportAllocs()
	for b.Loop() {
		rd.Reset(gz)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("статус = %d, ожидается %d", rr.Code, http.StatusOK)
		}
	}
}
