package common

import (
	"errors"
	"fmt"
	"sync"
)

type Topic int

func (t Topic) String() string {
	switch t {
	case JoinTopic:
		return "JoinTopic"
	case PingPongTopic:
		return "PingPongTopic"
	case PoolStepsTopic:
		return "PoolDataTopic"
	case PoolDataTopic:
		return "PoolDataTopic"
	default:
		return fmt.Sprintf("%d (cant find string)", int(t))
	}
}

const (
	JoinTopic Topic = iota + 1
	PingPongTopic
	PoolStepsTopic
	PoolDataTopic
)

var (
	UnknownMessageTypeErr = errors.New("Could not correlate server response to topic")
)

// TODO: find a way to not use interface{} for the channel
//       - Already boned on this once
type Broker struct {
	done        chan struct{}
	pubStream   chan interface{}
	subStream   chan topicSubscription
	unsubStream chan chan interface{}
	subs        map[Topic][]chan interface{}
	subCount    int
	mu          *sync.Mutex
}

func NewBroker(done chan struct{}) (b *Broker) {
	b = new(Broker)
	b.done = done
	b.pubStream = make(chan interface{}, 0)
	b.subStream = make(chan topicSubscription, 0)
	b.unsubStream = make(chan chan interface{}, 0)
	b.subs = make(map[Topic][]chan interface{})
	b.subs[JoinTopic] = make([]chan interface{}, 0)
	b.subs[PingPongTopic] = make([]chan interface{}, 0)
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
			// TODO: should I close every sub channel?
			return
		case sub := <-b.subStream:
			b.mu.Lock()
			b.subs[sub.topic] = append(b.subs[sub.topic], sub.subStream)
			b.mu.Unlock()
		case unsubStream := <-b.unsubStream:
			// subscriptions are always 1:1, never 1:many, so we can
			// iterate through entire map to find our sub and delete
			// it
			// TODO: If we dont find a sub, return an error
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
			// TODO: Need to groom out dead subscriptions?
			topics, err := findTopics(msg)
			if err != nil {
				// TODO: Better way to do this
				fmt.Println("Could not correlate server response to a topic: ", topics, err)
			}

			for _, topic := range topics {
				for _, stream := range b.subs[topic] {
					// fmt.Printf("Publishing %v to topic %v\n", msg, topic)
					stream <- msg
				}
			}
		}
	}
}

func findTopics(msg interface{}) ([]Topic, error) {
	switch msg.(type) {
	case joinOk:
		return []Topic{JoinTopic, PoolDataTopic}, nil
	case passFailed:
		return []Topic{JoinTopic}, nil
	case pong:
		return []Topic{PingPongTopic, PoolDataTopic}, nil
	case poolSteps:
		return []Topic{PoolStepsTopic, PoolDataTopic}, nil
	default:
		// TODO: Rethink how I'm doing this
		fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		fmt.Printf("Unknown message type: %s\n", msg)
		fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		return []Topic{}, UnknownMessageTypeErr
	}
}

func (b *Broker) Publish(msg interface{}) {
	b.pubStream <- msg
}

type topicSubscription struct {
	topic     Topic
	subStream chan interface{}
}

func (b *Broker) Subscribe(topic Topic) chan interface{} {
	subStream := make(chan interface{}, 1)
	t := topicSubscription{
		topic:     topic,
		subStream: subStream,
	}
	b.subStream <- t

	return subStream
}

func (b *Broker) Unsubscribe(unsub chan interface{}) {
	// TODO: How can I notify the caller that this failed?
	//       Maybe create and return an err channel, and pass
	//       err channel to Start?
	b.unsubStream <- unsub
}

func (b *Broker) SubscriptionCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	subCount := 0
	for _, v := range b.subs {
		subCount += len(v)
	}
	return subCount
}

// TODO: Can probably delete this
// type Subscription struct {
// 	done chan struct{}
// }
