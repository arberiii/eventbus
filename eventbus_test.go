package eventbus_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/short-d/eventbus"
)

func TestEventBusSimple(t *testing.T) {
	bus := eventbus.NewEventBus()

	notificationChannel := make(eventbus.DataChannel)
	topic := "greetings"
	bus.Subscribe(topic, notificationChannel)
	data := "Hello!"

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case result := <-notificationChannel:
				if result != data {
					t.Fatalf("expected data to be %v, but got %v", data, result)
				}
				wg.Done()
			}
		}
	}()

	bus.Publish(topic, data)

	if waitTimeout(&wg, 100*time.Millisecond) {
		t.Fatal("expected data to be published, but nothing was published")
	}
}

func TestUnSubscribe(t *testing.T) {
	bus := eventbus.NewEventBus()

	notificationChannel := make(eventbus.DataChannel)
	topic := "greetings"
	bus.Subscribe(topic, notificationChannel)
	subscribed := true
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case <-notificationChannel:
				if subscribed {
					bus.UnSubscribe(topic, notificationChannel)
					subscribed = false
					wg.Done()
				} else {
					t.Fatal("subscriber has unsubscribed but still got events")
				}
			}
		}
	}()

	data := "One"
	bus.Publish(topic, data)
	data = "Two"
	bus.Publish(topic, data)
	data = "Third"
	bus.Publish(topic, data)

	if waitTimeout(&wg, 100*time.Millisecond) {
		t.Fatal("expected data to be published, but nothing was published")
	}
}

func TestEventBusMultiple(t *testing.T) {
	numSubs := 100
	numPubs := 10
	bus := eventbus.NewEventBus()
	topic := "greetings"
	subscribers := make([]eventbus.DataChannel, numSubs)
	for i := 0; i < numSubs; i++ {
		subscribers[i] = make(eventbus.DataChannel)
		// subscribe to topic: topic + index mod numPubs
		bus.Subscribe(topic+strconv.Itoa(i%numPubs), subscribers[i])
	}
	data := "Hello"

	var wg sync.WaitGroup
	wg.Add(numSubs)

	for i := 0; i < numSubs; i++ {
		go func(i int) {
			for {
				select {
				case result := <-subscribers[i]:
					if result != data+strconv.Itoa(i%numPubs) {
						t.Fatalf("expected data to be %v, but got %v", data, result)
					}
					wg.Done()
				}
			}
		}(i)
	}
	for i := 0; i < numPubs; i++ {
		go func(i int) {
			bus.Publish(topic+strconv.Itoa(i%numPubs), data+strconv.Itoa(i%numPubs))
		}(i)
	}

	if waitTimeout(&wg, 100*time.Millisecond) {
		t.Fatal("expected data to be published, but some data were not published")
	}
}

// waitTimeout returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
