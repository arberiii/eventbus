package eventbus

import (
	"sync"
)

type Data interface{}
type DataChannel chan Data

type EventBus struct {
	mutex  sync.RWMutex
	events map[string][]DataChannel
}

func (e EventBus) Subscribe(eventName string, ch DataChannel) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if subscribers, ok := e.events[eventName]; ok {
		e.events[eventName] = append(subscribers, ch)
		return
	}

	e.events[eventName] = []DataChannel{ch}
}

func (e EventBus) UnSubscribe(eventName string, ch DataChannel) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	subscribers, ok := e.events[eventName]
	if !ok {
		return
	}

	for idx, subscriber := range subscribers {
		if subscriber == ch {
			subscribers[idx], subscribers[len(subscribers)-1] = subscribers[len(subscribers)-1], nil
			subscribers = subscribers[:len(subscribers)-1]

			if len(subscribers) == 0 {
				delete(e.events, eventName)
			}

			// drain the channel
			for {
				select {
				case <-ch:
				default:
					return
				}
			}
		}
	}
}

func (e EventBus) Publish(eventName string, data Data) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	subscribers, ok := e.events[eventName]
	if !ok {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(subscribers))
	for _, subscriber := range subscribers {
		subscriber <- data
		wg.Done()
	}
	wg.Wait()
}

func NewEventBus() EventBus {
	return EventBus{
		events: make(map[string][]DataChannel),
	}
}
