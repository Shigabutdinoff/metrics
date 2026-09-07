package persistent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shigabutdinoff/metrics/internal/storage"
	"go.uber.org/zap"
)

func fillStorage(st *storage.MemStorage, n int) {
	ctx := context.Background()
	for i := range n {
		v := float64(i) + 0.5
		st.SetGauge(ctx, fmt.Sprintf("Gauge%02d", i), &v)
	}
	delta := int64(42)
	st.AddCounter(ctx, "PollCount", &delta)
}

func TestService_SaveLoad(t *testing.T) {
	src := storage.NewMemStorage()
	fillStorage(src, 3)
	path := filepath.Join(t.TempDir(), "nested", "metrics.json")

	if err := New(src, path, zap.NewNop()).Save(); err != nil {
		t.Fatalf("Save() ошибка = %v", err)
	}

	dst := storage.NewMemStorage()
	if err := New(dst, path, zap.NewNop()).Load(); err != nil {
		t.Fatalf("Load() ошибка = %v", err)
	}

	ctx := context.Background()
	if g := dst.GetGauges(ctx)["Gauge01"]; g == nil || *g != 1.5 {
		t.Fatalf("gauge Gauge01 = %v, ожидается 1.5", g)
	}
	if c := dst.GetCounters(ctx)["PollCount"]; c == nil || *c != 42 {
		t.Fatalf("counter PollCount = %v, ожидается 42", c)
	}
}

func TestService_Load_NotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.json")
	if err := New(storage.NewMemStorage(), path, zap.NewNop()).Load(); err != nil {
		t.Fatalf("Load() для отсутствующего файла ошибка = %v, ожидается nil", err)
	}
}

func TestService_Load_BadJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := New(storage.NewMemStorage(), path, zap.NewNop()).Load(); err == nil {
		t.Fatal("Load() для битого JSON вернул nil, ожидается ошибка")
	}
}

func BenchmarkService_Save(b *testing.B) {
	st := storage.NewMemStorage()
	fillStorage(st, 45)
	s := New(st, filepath.Join(b.TempDir(), "metrics.json"), zap.NewNop())

	b.ReportAllocs()
	for b.Loop() {
		if err := s.Save(); err != nil {
			b.Fatalf("Save() ошибка = %v", err)
		}
	}
}
