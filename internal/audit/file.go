package audit

import (
	"context"
	"encoding/json"
	"os"
)

const filePerm = 0o600

// FileSink дописывает события аудита в файл по одному на строку.
type FileSink struct {
	f   *os.File
	enc *json.Encoder
}

// NewFileSink открывает файл аудита на дозапись, создавая его.
func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return nil, err
	}

	return &FileSink{f: f, enc: json.NewEncoder(f)}, nil
}

// Update дописывает событие в файл, отменённый ctx прекращает запись.
func (s *FileSink) Update(ctx context.Context, e Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.enc.Encode(e)
}

// Close закрывает файл аудита.
func (s *FileSink) Close() error {
	return s.f.Close()
}
