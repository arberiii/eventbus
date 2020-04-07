package eventbus_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/short-d/eventbus"
)

func TestEventBusPub(t *testing.T) {
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

func TestEventBusPubAsync(t *testing.T) {
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
			bus.PublishAsync(topic+strconv.Itoa(i%numPubs), data+strconv.Itoa(i%numPubs))
		}(i)
	}

	if waitTimeout(&wg, 100*time.Millisecond) {
		t.Fatal("expected data to be published, but some data were not published")
	}
}

func TestUnSubscribe(t *testing.T) {
	bus := eventbus.NewEventBus()

	s1 := make(eventbus.DataChannel)
	topic := "greetings"
	bus.Subscribe(topic, s1)
	subscribed := true
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case <-s1:
				if subscribed {
					bus.UnSubscribe(topic, s1)
					subscribed = false
					wg.Done()
				} else {
					t.Fatal("subscriber has unsubscribed but still got events")
				}
			}
		}
	}()

	s2 := make(eventbus.DataChannel)
	bus.Subscribe(topic, s2)
	wg.Add(3)

	go func() {
		for {
			select {
			case <-s2:
				wg.Done()
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
