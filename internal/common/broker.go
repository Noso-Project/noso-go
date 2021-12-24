package common

import (
	"fmt"
	"sync"
)

type Broker struct {
	done        chan struct{}
	pubStream   chan interface{}
	subStream   chan chan interface{}
	unsubStream chan chan interface{}
	subs        map[ServerMessageType][]chan interface{}
}

func NewBroker(done chan struct{}) (b *Broker) {
	b = new(Broker)
	b.done = done
	b.pubStream = make(chan interface{}, 0)
	b.subStream = make(chan chan interface{}, 0)
	b.unsubStream = make(chan chan interface{}, 0)
	b.subs = make(map[ServerMessageType][]chan interface{})
	b.subs[JOINOK] = make([]chan interface{}, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go b.start(&wg)
	wg.Wait()

	return
}

func (b *Broker) start(wg *sync.WaitGroup) {
	wg.Done()
	for {
		select {
		case <-b.done:
			return
		case subStream := <-b.subStream:
			b.subs[JOINOK] = append(b.subs[JOINOK], subStream)
		case msg := <-b.pubStream:
			fmt.Println("From pubStream: ", msg)
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
	fmt.Println("From Publish: ", msg)
	b.pubStream <- msg
}

func (b *Broker) Subscribe(msgType ServerMessageType) chan interface{} {
	subCh := make(chan interface{}, 1)
	b.subStream <- subCh

	return subCh
}
