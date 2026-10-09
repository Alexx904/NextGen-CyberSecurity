package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
)

func TestBroadcast_SingleSubscriber(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	topic := "scan:test-single"
	ch, unsub := broadcaster.Subscribe(context.Background(), topic)
	defer unsub()

	scanID := uuid.New()
	event := domain.SSEEvent{
		Event: domain.EventScanProgress,
		Data: domain.ScanProgressPayload{
			ScanID:     scanID,
			Percentage: 42.5,
			Message:    "SYN Stealth Scan in progress",
		},
	}

	broadcaster.Publish(topic, event)

	select {
	case received, ok := <-ch:
		if !ok {
			t.Fatal("il canale è stato chiuso inaspettatamente")
		}
		if received.Event != domain.EventScanProgress {
			t.Fatalf("atteso tipo evento %s, ottenuto %s", domain.EventScanProgress, received.Event)
		}
		payload, ok := received.Data.(domain.ScanProgressPayload)
		if !ok {
			t.Fatalf("tipo payload non corretto: %T", received.Data)
		}
		if payload.ScanID != scanID || payload.Percentage != 42.5 {
			t.Fatalf("payload non corrispondente: %+v", payload)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout in attesa dell'evento broadcast")
	}
}

func TestBroadcast_MultipleSubscribers(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	topic := "discovery:multi"
	ch1, unsub1 := broadcaster.Subscribe(context.Background(), topic)
	defer unsub1()
	ch2, unsub2 := broadcaster.Subscribe(context.Background(), topic)
	defer unsub2()
	ch3, unsub3 := broadcaster.Subscribe(context.Background(), topic)
	defer unsub3()

	if count := broadcaster.SubscriberCount(topic); count != 3 {
		t.Fatalf("attesi 3 subscriber, trovati %d", count)
	}

	sessionID := uuid.New()
	event := domain.SSEEvent{
		Event: domain.EventDeviceDiscovered,
		Data: domain.DeviceDiscoveredPayload{
			SessionID: sessionID,
			Device: domain.DiscoveredDevice{
				IP:     "192.168.1.50",
				Vendor: "Raspberry Pi",
				Status: "up",
			},
		},
	}

	broadcaster.Publish(topic, event)

	channels := []<-chan domain.SSEEvent{ch1, ch2, ch3}
	for i, ch := range channels {
		select {
		case received := <-ch:
			if received.Event != domain.EventDeviceDiscovered {
				t.Fatalf("subscriber %d: atteso evento %s, ottenuto %s", i, domain.EventDeviceDiscovered, received.Event)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("subscriber %d: timeout in attesa dell'evento", i)
		}
	}
}

func TestBroadcast_TopicIsolation(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	topicA := "scan:topic-a"
	topicB := "scan:topic-b"

	chA, unsubA := broadcaster.Subscribe(context.Background(), topicA)
	defer unsubA()
	chB, unsubB := broadcaster.Subscribe(context.Background(), topicB)
	defer unsubB()

	eventA := domain.SSEEvent{
		Event: domain.EventScanPhase,
		Data: domain.ScanPhasePayload{
			Phase: "PORT_SCAN",
		},
	}

	broadcaster.Publish(topicA, eventA)

	// Canale A deve ricevere l'evento
	select {
	case received := <-chA:
		if received.Event != domain.EventScanPhase {
			t.Fatalf("atteso %s, ottenuto %s", domain.EventScanPhase, received.Event)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("subscriber A non ha ricevuto l'evento")
	}

	// Canale B non deve ricevere nulla
	select {
	case event := <-chB:
		t.Fatalf("subscriber B ha ricevuto un evento non pertinente: %+v", event)
	case <-time.After(50 * time.Millisecond):
		// Corretto: nessun evento su topicB
	}
}

func TestBroadcast_SubscriberCount(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	topic := "stats:topic"
	if count := broadcaster.SubscriberCount(topic); count != 0 {
		t.Fatalf("atteso 0 per topic inesistente, trovato %d", count)
	}

	_, unsub1 := broadcaster.Subscribe(context.Background(), topic)
	_, unsub2 := broadcaster.Subscribe(context.Background(), topic)

	if count := broadcaster.SubscriberCount(topic); count != 2 {
		t.Fatalf("attesi 2 subscriber, trovati %d", count)
	}

	unsub1()
	if count := broadcaster.SubscriberCount(topic); count != 1 {
		t.Fatalf("atteso 1 subscriber dopo unsub1, trovato %d", count)
	}

	unsub2()
	if count := broadcaster.SubscriberCount(topic); count != 0 {
		t.Fatalf("atteso 0 subscriber dopo unsub2, trovato %d", count)
	}
}

func TestBroadcast_PublishNoSubscribers(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	// Pubblicazione su topic senza subscriber: nessuna eccezione o blocco
	broadcaster.Publish("empty:topic", domain.SSEEvent{
		Event: "test",
		Data:  "nobody is listening",
	})
}

func TestBroadcast_IdempotentUnsubscribe(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	topic := "idempotent:unsub"
	ch, unsub := broadcaster.Subscribe(context.Background(), topic)

	// Chiamata multipla a unsub() non deve causare panic da double close
	unsub()
	unsub()
	unsub()

	// Il canale deve essere chiuso
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("il canale doveva essere chiuso")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout in attesa della chiusura del canale")
	}
}

func TestBroadcast_ContextCancellation(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)
	defer broadcaster.Close()

	ctx, cancel := context.WithCancel(context.Background())
	topic := "ctx:cancel"

	ch, _ := broadcaster.Subscribe(ctx, topic)
	if count := broadcaster.SubscriberCount(topic); count != 1 {
		t.Fatalf("atteso 1 subscriber, trovato %d", count)
	}

	cancel()

	// Alla cancellazione del context, il canale deve chiudersi e il subscriber essere rimosso
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("il canale doveva essere chiuso a seguito di ctx.Cancel()")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout in attesa della chiusura automatica del canale")
	}

	// Breve attesa per consentire la sincronizzazione della mappa
	time.Sleep(20 * time.Millisecond)
	if count := broadcaster.SubscriberCount(topic); count != 0 {
		t.Fatalf("atteso 0 subscriber dopo ctx.Cancel(), trovato %d", count)
	}
}

func TestBroadcast_SlowSubscriberDropPolicy(t *testing.T) {
	// Buffer piccolissimo di 2 elementi per forzare l'overflow
	broadcaster := NewEventBroadcaster(2, nil)
	defer broadcaster.Close()

	topic := "slow:sub"
	ch, unsub := broadcaster.Subscribe(context.Background(), topic)
	defer unsub()

	// Invio di 10 eventi consecutivi senza consumarli
	for i := 1; i <= 10; i++ {
		broadcaster.Publish(topic, domain.SSEEvent{
			Event: "number",
			Data:  i,
		})
	}

	// Il canale ha capienza 2: deve contenere i primi 2 eventi
	receivedCount := 0
drainLoop:
	for {
		select {
		case <-ch:
			receivedCount++
		default:
			break drainLoop
		}
	}

	if receivedCount != 2 {
		t.Fatalf("attesi esattamente 2 eventi nel buffer saturo, ricevuti %d", receivedCount)
	}
}

func TestBroadcast_CloseBroadcaster(t *testing.T) {
	broadcaster := NewEventBroadcaster(10, nil)

	topic := "close:test"
	ch1, _ := broadcaster.Subscribe(context.Background(), topic)
	ch2, _ := broadcaster.Subscribe(context.Background(), topic)

	broadcaster.Close()

	// Nuova chiusura: non deve causare panic
	broadcaster.Close()

	// I canali esistenti devono essere chiusi
	for i, ch := range []<-chan domain.SSEEvent{ch1, ch2} {
		select {
		case _, ok := <-ch:
			if ok {
				t.Fatalf("canale %d doveva essere chiuso", i)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout chiusura canale %d", i)
		}
	}

	// Nuova iscrizione dopo Close(): deve restituire canale già chiuso
	chNew, unsubNew := broadcaster.Subscribe(context.Background(), topic)
	unsubNew()
	select {
	case _, ok := <-chNew:
		if ok {
			t.Fatal("canale iscritto dopo Close() doveva essere già chiuso")
		}
	default:
	}

	// Nuova Publish dopo Close(): no-op
	broadcaster.Publish(topic, domain.SSEEvent{Event: "test"})
}

func TestFormatSSE(t *testing.T) {
	tests := []struct {
		name     string
		event    domain.SSEEvent
		expected []string
	}{
		{
			name: "Completo con ID e Retry",
			event: domain.SSEEvent{
				ID:    "evt-123",
				Event: domain.EventScanPhase,
				Retry: 5000,
				Data: map[string]string{
					"phase": "PORT_SCAN",
				},
			},
			expected: []string{
				"id: evt-123\n",
				"event: phase\n",
				"retry: 5000\n",
				"data: {\"phase\":\"PORT_SCAN\"}\n\n",
			},
		},
		{
			name: "Senza ID e Retry",
			event: domain.SSEEvent{
				Event: domain.EventDeviceDiscovered,
				Data: domain.DiscoveredDevice{
					IP:     "192.168.1.1",
					Status: "up",
				},
			},
			expected: []string{
				"event: device_discovered\n",
				"data: {",
				"\"ip\":\"192.168.1.1\"",
				"\"status\":\"up\"",
				"\n\n",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.event.FormatBytes()
			if err != nil {
				t.Fatalf("FormatBytes ha restituito un errore inaspettato: %v", err)
			}
			out := string(b)
			for _, exp := range tc.expected {
				if !strings.Contains(out, exp) {
					t.Fatalf("output non contiene %q:\n%s", exp, out)
				}
			}
		})
	}
}

func TestBroadcast_ConcurrentPublishSubscribe(t *testing.T) {
	broadcaster := NewEventBroadcaster(32, nil)
	defer broadcaster.Close()

	const (
		numPublishers  = 20
		numSubscribers = 20
		eventsPerPub   = 50
	)

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Launch concurrent subscribers
	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(subIdx int) {
			defer wg.Done()
			topic := fmt.Sprintf("topic-%d", subIdx%5)
			ch, unsub := broadcaster.Subscribe(ctx, topic)
			defer unsub()

			count := 0
			for {
				select {
				case _, ok := <-ch:
					if !ok {
						return
					}
					count++
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// Launch concurrent publishers
	for i := 0; i < numPublishers; i++ {
		wg.Add(1)
		go func(pubIdx int) {
			defer wg.Done()
			topic := fmt.Sprintf("topic-%d", pubIdx%5)
			for j := 0; j < eventsPerPub; j++ {
				broadcaster.Publish(topic, domain.SSEEvent{
					Event: "stress_event",
					Data:  fmt.Sprintf("pub-%d-msg-%d", pubIdx, j),
				})
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}

