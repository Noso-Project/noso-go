package common

import (
	"fmt"
	"sync"
)

// TODO: find a way to not use interface{} for the channel
type Broker struct {
	done        chan struct{}
	pubStream   chan interface{}
	subStream   chan chan interface{}
	unsubStream chan chan interface{}
	subs        map[ServerMessageType][]chan interface{}
	subCount    int
	mu          *sync.Mutex
}

func NewBroker(done chan struct{}) (b *Broker) {
	b = new(Broker)
	b.done = done
	b.pubStream = make(chan interface{}, 0)
	b.subStream = make(chan chan interface{}, 0)
	b.unsubStream = make(chan chan interface{}, 0)
	b.subs = make(map[ServerMessageType][]chan interface{})
	b.subs[JOINOK] = make([]chan interface{}, 0)
	b.subCount = 0
	b.mu = new(sync.Mutex)

	var wg sync.WaitGroup
	wg.Add(1)
	go b.start(&wg)
	wg.Wait()

	return
}

func removeIndex(s []chan interface{}, index int) []chan interface{} {
	return append(s[:index], s[index+1:]...)
}

func (b *Broker) start(wg *sync.WaitGroup) {
	wg.Done()
	for {
		select {
		case <-b.done:
			return
		case subStream := <-b.subStream:
			fmt.Println("Subscribing to: ", subStream)
			b.mu.Lock()
			b.subs[JOINOK] = append(b.subs[JOINOK], subStream)
			b.mu.Unlock()
		case unsubStream := <-b.unsubStream:
			for k, v := range b.subs {
				for idx, ch := range v[:] {
					if ch == unsubStream {
						b.mu.Lock()
						b.subs[k] = removeIndex(b.subs[k], idx)
						b.mu.Unlock()
					}
				}
			}
			close(unsubStream)
		case msg := <-b.pubStream:
			// fmt.Println("From pubStream: ", msg)
			var msgType ServerMessageType
			switch msg.(type) {
			case joinOk:
				msgType = msg.(joinOk).MsgType
			default:
				// TODO: should log error here?
				fmt.Printf("Unknown event type: %v\n", msg)
				continue
			}
			for _, msgCh := range b.subs[msgType] {
				msgCh <- msg
			}
		}
	}
}

func (b *Broker) Publish(msg interface{}) {
	// fmt.Println("From Publish: ", msg)
	b.pubStream <- msg
}

func (b *Broker) Subscribe(msgType ServerMessageType) chan interface{} {
	subCh := make(chan interface{}, 1)
	b.subStream <- subCh

	return subCh
}

func (b *Broker) Unsubscribe(unsub chan interface{}) {
	// TODO: How can I notify the caller that this failed?
	b.unsubStream <- unsub
}

func (b *Broker) SubscriptionCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	subCount := 0
	for _, v := range b.subs {
		fmt.Println("Subcount is: ", subCount)
		subCount += len(v)
	}
	return subCount
}

type Subscription struct {
	done chan struct{}
}
