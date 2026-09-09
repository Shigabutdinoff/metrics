package audit

import (
	"context"
	"sync"
)

type recordKey struct{}

// Names потокобезопасно собирает имена метрик запроса для события аудита.
type Names struct {
	mu    sync.Mutex
	names []string
}

// Collected возвращает копию собранных имён.
func (n *Names) Collected() []string {
	n.mu.Lock()
	defer n.mu.Unlock()

	return append([]string(nil), n.names...)
}

func (n *Names) add(names ...string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.names = append(n.names, names...)
}

// WithRecord готовит контекст к сбору имён метрик для события аудита.
func WithRecord(ctx context.Context, rec *Names) context.Context {
	return context.WithValue(ctx, recordKey{}, rec)
}

// Record добавляет имена принятых метрик в событие аудита запроса.
func Record(ctx context.Context, names ...string) {
	rec, ok := ctx.Value(recordKey{}).(*Names)
	if !ok {
		return
	}
	rec.add(names...)
}
