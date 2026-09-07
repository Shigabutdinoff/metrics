package reqbody

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

const maxBodySize = 10 << 20

type body struct {
	bytes.Reader
	data []byte
}

func (*body) Close() error { return nil }

// Read читает тело с лимитом 10 МБ и возвращает его же в r.Body.
func Read(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	// повторный вызов из следующего middleware отдаёт уже прочитанное тело
	if b, ok := r.Body.(*body); ok {
		b.Reset(b.data)
		return b.data, true
	}
	if r.Body == nil || r.Body == http.NoBody {
		return nil, true
	}

	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "тело запроса слишком велико", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "не удалось прочитать тело запроса", http.StatusBadRequest)
		}
		return nil, false
	}
	b := &body{data: data}
	b.Reset(data)
	r.Body = b

	return data, true
}
