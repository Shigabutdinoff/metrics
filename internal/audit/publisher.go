package audit

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultBuffer       = 256
	defaultCloseTimeout = 5 * time.Second
	hardStopGrace       = 100 * time.Millisecond
)

// Observer приёмник событий аудита (подписчик).
type Observer interface {
	Update(ctx context.Context, e Event) error
}

// Notifier издатель с точки зрения публикующего кода.
type Notifier interface {
	Publish(e Event)
}

type subscription struct {
	obs  Observer
	name string
	ch   chan Event
}

// Publisher асинхронно рассылает события аудита наблюдателям.
type Publisher struct {
	mu           sync.RWMutex
	subs         []*subscription
	wg           sync.WaitGroup
	buffer       int
	log          *zap.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	closeTimeout time.Duration
	closed       bool
}

// Option настраивает издателя при создании.
type Option func(*Publisher)

// WithBuffer задаёт размер очереди событий у каждого приёмника.
func WithBuffer(n int) Option {
	return func(p *Publisher) { p.buffer = n }
}

// WithCloseTimeout задаёт предел ожидания разборщиков при Close.
func WithCloseTimeout(d time.Duration) Option {
	return func(p *Publisher) { p.closeTimeout = d }
}

// NewPublisher создаёт издателя с буферизованной очередью у каждого приёмника.
func NewPublisher(log *zap.Logger, opts ...Option) *Publisher {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Publisher{
		buffer:       defaultBuffer,
		log:          log,
		ctx:          ctx,
		cancel:       cancel,
		closeTimeout: defaultCloseTimeout,
	}
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Register подписывает наблюдателя и запускает разборщик его очереди.
func (p *Publisher) Register(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}

	s := &subscription{
		obs:  o,
		name: fmt.Sprintf("%T", o),
		ch:   make(chan Event, p.buffer),
	}
	p.subs = append(p.subs, s)
	p.wg.Go(func() { p.consume(s) })
}

// Deregister отписывает наблюдателя; несравнимый тип отписать нельзя.
func (p *Publisher) Deregister(o Observer) {
	if o == nil || !reflect.TypeOf(o).Comparable() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}

	for i, s := range p.subs {
		if s.obs == o {
			p.subs = append(p.subs[:i], p.subs[i+1:]...)
			close(s.ch)
			return
		}
	}
}

// Publish кладёт событие в очередь каждого приёмника без блокировки.
func (p *Publisher) Publish(e Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return
	}

	for _, s := range p.subs {
		select {
		case s.ch <- e:
		default:
			p.log.Warn("Буфер аудита переполнен, событие отброшено",
				zap.String("observer", s.name),
			)
		}
	}
}

// Close останавливает разборщики очередей, наблюдателей не закрывает.
func (p *Publisher) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	for _, s := range p.subs {
		close(s.ch)
	}
	p.mu.Unlock()

	p.drain()
	p.cancel()
}

func (p *Publisher) drain() {
	drained := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(drained)
	}()

	timer := time.NewTimer(p.closeTimeout)
	defer timer.Stop()

	select {
	case <-drained:
		return
	case <-timer.C:
		p.cancel()
		p.log.Warn("Приёмник аудита не остановился, ожидание прекращено",
			zap.Duration("waited", p.closeTimeout),
		)
	}

	grace := time.NewTimer(hardStopGrace)
	defer grace.Stop()

	select {
	case <-drained:
	case <-grace.C:
	}
}

func (p *Publisher) consume(s *subscription) {
	var dropped uint64
	for e := range s.ch {
		if p.ctx.Err() != nil {
			dropped++
			continue
		}
		err := s.obs.Update(p.ctx, e)
		if err == nil {
			continue
		}
		if p.ctx.Err() != nil {
			dropped++
			continue
		}
		p.log.Warn("Не удалось отправить событие аудита",
			zap.String("observer", s.name),
			zap.Error(err),
		)
	}
	if dropped > 0 {
		p.log.Warn("События аудита отброшены после таймаута Close",
			zap.String("observer", s.name),
			zap.Uint64("dropped", dropped),
		)
	}
}
