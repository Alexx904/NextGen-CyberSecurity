package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"nextgen-cybersec/internal/domain"
)

// DefaultEventBufferSize definisce la capacità standard del buffer per ciascun subscriber
const DefaultEventBufferSize = 64

// EventBroadcaster definisce il contratto per il pub/sub in-memory per lo streaming SSE
type EventBroadcaster interface {
	// Subscribe crea un canale per il topic specificato. Restituisce il canale di ricezione
	// e una funzione di disiscrizione (idempotente). Se il context viene cancellato,
	// l'unsubscription avviene automaticamente.
	Subscribe(ctx context.Context, topic string) (<-chan domain.SSEEvent, func())

	// Publish invia un evento a tutti i subscriber registrati sul topic specificato.
	// L'invio è non bloccante: se il buffer di un subscriber è saturo, l'evento viene
	// scartato con warning di log (drop policy) per preservare le prestazioni del sistema.
	Publish(topic string, event domain.SSEEvent)

	// SubscriberCount restituisce il numero di subscriber attualmente registrati su un topic.
	SubscriberCount(topic string) int

	// Close arresta il broadcaster e chiude in modo sicuro tutti i canali dei subscriber attivi.
	Close()
}

type subscriber struct {
	id     uint64
	topic  string
	ch     chan domain.SSEEvent
	once   sync.Once
	closed atomic.Bool
}

type inMemoryEventBroadcaster struct {
	mu         sync.RWMutex
	topics     map[string]map[uint64]*subscriber
	bufferSize int
	logger     *slog.Logger
	nextSubID  atomic.Uint64
	closed     atomic.Bool
}

// NewEventBroadcaster istanzia un nuovo EventBroadcaster in-memory thread-safe
func NewEventBroadcaster(bufferSize int, logger *slog.Logger) EventBroadcaster {
	if bufferSize <= 0 {
		bufferSize = DefaultEventBufferSize
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &inMemoryEventBroadcaster{
		topics:     make(map[string]map[uint64]*subscriber),
		bufferSize: bufferSize,
		logger:     logger,
	}
}

// Subscribe registra un nuovo subscriber per il topic specificato
func (b *inMemoryEventBroadcaster) Subscribe(ctx context.Context, topic string) (<-chan domain.SSEEvent, func()) {
	if b.closed.Load() {
		closedCh := make(chan domain.SSEEvent)
		close(closedCh)
		return closedCh, func() {}
	}

	subID := b.nextSubID.Add(1)
	ch := make(chan domain.SSEEvent, b.bufferSize)
	sub := &subscriber{
		id:    subID,
		topic: topic,
		ch:    ch,
	}

	b.mu.Lock()
	if b.closed.Load() {
		b.mu.Unlock()
		close(ch)
		return ch, func() {}
	}

	if b.topics[topic] == nil {
		b.topics[topic] = make(map[uint64]*subscriber)
	}
	b.topics[topic][subID] = sub
	b.mu.Unlock()

	unsubscribe := func() {
		sub.once.Do(func() {
			sub.closed.Store(true)

			b.mu.Lock()
			if subs, ok := b.topics[topic]; ok {
				delete(subs, subID)
				if len(subs) == 0 {
					delete(b.topics, topic)
				}
			}
			b.mu.Unlock()

			close(sub.ch)
		})
	}

	if ctx != nil {
		go func() {
			select {
			case <-ctx.Done():
				unsubscribe()
			}
		}()
	}

	return sub.ch, unsubscribe
}

// Publish distribuisce l'evento a tutti i subscriber del topic indicato senza bloccare il chiamante
func (b *inMemoryEventBroadcaster) Publish(topic string, event domain.SSEEvent) {
	if b.closed.Load() {
		return
	}

	b.mu.RLock()
	subsMap, ok := b.topics[topic]
	if !ok || len(subsMap) == 0 {
		b.mu.RUnlock()
		return
	}

	// Copia locale dei subscriber per minimizzare il tempo di detenzione del lock
	subs := make([]*subscriber, 0, len(subsMap))
	for _, sub := range subsMap {
		subs = append(subs, sub)
	}
	b.mu.RUnlock()

	for _, sub := range subs {
		if sub.closed.Load() {
			continue
		}

		select {
		case sub.ch <- event:
		default:
			// Drop policy non-bloccante: notifica di congestione
			b.logger.Warn("Canale SSE saturo, evento scartato per subscriber",
				slog.String("topic", topic),
				slog.Uint64("subscriber_id", sub.id),
				slog.String("event", event.Event),
			)
		}
	}
}

// SubscriberCount restituisce il conteggio attuale di listener per un topic
func (b *inMemoryEventBroadcaster) SubscriberCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.topics[topic])
}

// Close arresta il broadcaster e rilascia tutti i subscriber
func (b *inMemoryEventBroadcaster) Close() {
	if b.closed.Swap(true) {
		return
	}

	b.mu.Lock()
	allSubs := make([]*subscriber, 0)
	for _, subsMap := range b.topics {
		for _, sub := range subsMap {
			allSubs = append(allSubs, sub)
		}
	}
	b.topics = make(map[string]map[uint64]*subscriber)
	b.mu.Unlock()

	for _, sub := range allSubs {
		sub.once.Do(func() {
			sub.closed.Store(true)
			close(sub.ch)
		})
	}
}

