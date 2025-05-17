package streamhub

import (
	"context"
	"sync"
)

type Event interface{}

type StreamHub struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
	input       chan Event
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewStreamHub(bufferSize int) *StreamHub {
	ctx, cancel := context.WithCancel(context.Background())
	hub := &StreamHub{
		subscribers: make(map[chan Event]struct{}),
		input:       make(chan Event, bufferSize),
		ctx:         ctx,
		cancel:      cancel,
	}
	return hub
}

// Отдаёт канал, в который можно писать и который стрим разошлёт подписчикам
func (h *StreamHub) SubscribeAsProducer() chan<- Event {
	return h.input
}

// Возвращает канал, по которому можно получать события
func (h *StreamHub) Subscribe(bufferSize int) <-chan Event {
	ch := make(chan Event, bufferSize)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	// Автоматическое удаление при закрытии подписки
	go func() {
		<-h.ctx.Done()
		h.mu.Lock()
		delete(h.subscribers, ch)
		close(ch)
		h.mu.Unlock()
	}()

	return ch
}

// Остановка потока
func (h *StreamHub) Stop() {
	h.cancel()
}

// Основная логика доставки: каждое событие -> всем подписчикам
func (h *StreamHub) Start() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case event := <-h.input:
			h.mu.RLock()
			for ch := range h.subscribers {
				select {
				case ch <- event:
				default:
					// если буфер переполнен — пропускаем
					// можно заменить на drop, log, backpressure...
				}
			}
			h.mu.RUnlock()
		}
	}
}
