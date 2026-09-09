package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/shigabutdinoff/metrics/internal/model/metrics"
)

func TestMemStorage_AddCounter(t *testing.T) {
	ms := &MemStorage{}

	first := int64(2)
	second := int64(5)

	ctx := context.Background()
	ms.AddCounter(ctx, "requests", &first)
	ms.AddCounter(ctx, "requests", &second)

	got := ms.GetCounters(ctx)["requests"]
	if got == nil {
		t.Fatal("counter requests равен nil")
	}
	if *got != 7 {
		t.Fatalf("counter requests = %d, ожидается %d", *got, 7)
	}
}

func TestMemStorage_GetCounters(t *testing.T) {
	ms := &MemStorage{}
	got := ms.GetCounters(context.Background())

	if got == nil {
		t.Fatal("GetCounters() вернул nil map")
	}
	if len(got) != 0 {
		t.Fatalf("len(GetCounters()) = %d, ожидается 0", len(got))
	}
}

func TestMemStorage_GetGauges(t *testing.T) {
	ms := &MemStorage{}
	got := ms.GetGauges(context.Background())

	if got == nil {
		t.Fatal("GetGauges() вернул nil map")
	}
	if len(got) != 0 {
		t.Fatalf("len(GetGauges()) = %d, ожидается 0", len(got))
	}
}

func TestMemStorage_SetGauge(t *testing.T) {
	ms := &MemStorage{}

	v1 := 1.5
	v2 := 2.5
	ctx := context.Background()
	ms.SetGauge(ctx, "alloc", metrics.GaugeValue(&v1))
	ms.SetGauge(ctx, "alloc", metrics.GaugeValue(&v2))

	got := ms.GetGauges(ctx)["alloc"]
	if got == nil {
		t.Fatal("gauge alloc равен nil")
	}
	if *got != 2.5 {
		t.Fatalf("gauge alloc = %f, ожидается %f", *got, 2.5)
	}
}

func TestNewMemStorage(t *testing.T) {
	got := NewMemStorage()
	if got == nil {
		t.Fatal("NewMemStorage() вернул nil")
	}
	if got.gauges == nil {
		t.Fatal("gauges map равна nil")
	}
	if got.counters == nil {
		t.Fatal("counters map равна nil")
	}
	if len(got.gauges) != 0 || len(got.counters) != 0 {
		t.Fatalf("неожиданно непустые maps: gauges=%d counters=%d", len(got.gauges), len(got.counters))
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	ms := NewMemStorage()
	ctx := context.Background()
	v := 1.5
	ms.SetGauge(ctx, "alloc", &v)

	got := ms.GetGauge(ctx, "alloc")
	if got == nil || *got != 1.5 {
		t.Fatalf("GetGauge(alloc) = %v, ожидается 1.5", got)
	}
	if absent := ms.GetGauge(ctx, "absent"); absent != nil {
		t.Fatalf("GetGauge(absent) = %v, ожидается nil", absent)
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	ms := NewMemStorage()
	ctx := context.Background()
	delta := int64(7)
	ms.AddCounter(ctx, "requests", &delta)

	got := ms.GetCounter(ctx, "requests")
	if got == nil || *got != 7 {
		t.Fatalf("GetCounter(requests) = %v, ожидается 7", got)
	}
	*got = 100
	if again := ms.GetCounter(ctx, "requests"); *again != 7 {
		t.Fatalf("копия счётчика изменила хранилище: %d", *again)
	}
	if absent := ms.GetCounter(ctx, "absent"); absent != nil {
		t.Fatalf("GetCounter(absent) = %v, ожидается nil", absent)
	}
}

func fillStorage(ms *MemStorage, n int) {
	ctx := context.Background()
	for i := range n {
		v := float64(i) + 0.5
		ms.SetGauge(ctx, fmt.Sprintf("Gauge%02d", i), &v)
		delta := int64(i)
		ms.AddCounter(ctx, fmt.Sprintf("Counter%02d", i), &delta)
	}
}

func BenchmarkMemStorage_SetGauge(b *testing.B) {
	ms := NewMemStorage()
	ctx := context.Background()
	v := 2457600.5

	b.ReportAllocs()
	for b.Loop() {
		ms.SetGauge(ctx, "Alloc", &v)
	}
}

func BenchmarkMemStorage_AddCounter(b *testing.B) {
	ms := NewMemStorage()
	ctx := context.Background()
	delta := int64(1)

	b.ReportAllocs()
	for b.Loop() {
		ms.AddCounter(ctx, "PollCount", &delta)
	}
}

func benchGet(b *testing.B, get func()) {
	b.ReportAllocs()
	for b.Loop() {
		get()
	}
}

func BenchmarkMemStorage_GetAll(b *testing.B) {
	ms := NewMemStorage()
	fillStorage(ms, 45)
	ctx := context.Background()

	b.Run("gauges", func(b *testing.B) { benchGet(b, func() { ms.GetGauges(ctx) }) })
	b.Run("counters", func(b *testing.B) { benchGet(b, func() { ms.GetCounters(ctx) }) })
}

func BenchmarkMemStorage_GetOne(b *testing.B) {
	ms := NewMemStorage()
	fillStorage(ms, 45)
	ctx := context.Background()

	b.Run("gauge", func(b *testing.B) { benchGet(b, func() { ms.GetGauge(ctx, "Gauge10") }) })
	b.Run("counter", func(b *testing.B) { benchGet(b, func() { ms.GetCounter(ctx, "Counter10") }) })
}
