package events

import (
	"context"
	"sync"
)

// Event is a domain event marker.
// Name should be stable; used for routing to subscribers.
type Event interface {
	Name() string
}

type Handler func(context.Context, Event) error

// Bus is an in-process pub/sub event bus.
// It is intentionally simple: best-effort delivery to handlers in registration order.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{handlers: map[string][]Handler{}}
}

func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

func (b *Bus) Publish(ctx context.Context, evt Event) error {
	b.mu.RLock()
	hs := append([]Handler(nil), b.handlers[evt.Name()]...)
	b.mu.RUnlock()

	for _, h := range hs {
		if err := h(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

