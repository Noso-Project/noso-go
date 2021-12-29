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
	case StepOkTopic:
		return "StepOkTopic"
	default:
		return fmt.Sprintf("%d (cant find string)", int(t))
	}
}

const (
	JoinTopic Topic = iota + 1
	PingPongTopic
	PoolStepsTopic
	PoolDataTopic
	StepOkTopic
)

var (
	ErrUnknownMessageType = errors.New("Could not correlate server response to topic")
)

// TODO: find a way to not use interface{} for the channel
//       - Already boned on this once
type Broker struct {
	done        chan struct{}
	pubStream   chan interface{}
	subStream   chan topicSubscription
	unsubStream chan (<-chan interface{})
	subs        map[Topic][]chan interface{}
	subCount    int
	mu          *sync.Mutex
}

func NewBroker(done chan struct{}) (b *Broker) {
	b = new(Broker)
	b.done = done
	b.pubStream = make(chan interface{}, 0)
	b.subStream = make(chan topicSubscription, 0)
	b.unsubStream = make(chan (<-chan interface{}), 0)
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

// TODO: Need to add a stop method
func (b *Broker) start(wg *sync.WaitGroup) {
	wg.Done()
	for {
		select {
		case <-b.done:
			// Attempt to close every stream in subs
			b.mu.Lock()
			for _, v := range b.subs {
				for _, stream := range v[:] {
					close(stream)
				}
			}
			b.mu.Unlock()
			return
		case sub := <-b.subStream:
			// fmt.Println("22222222222222222222222222")
			// fmt.Printf("About to sub: topic %v channel %v\n", sub.topic, sub.subStream)
			// fmt.Println("22222222222222222222222222")
			b.mu.Lock()
			b.subs[sub.topic] = append(b.subs[sub.topic], sub.subStream)
			// fmt.Println("33333333333333333333333333")
			// fmt.Printf("Topics subs now: %v\n", b.subs[sub.topic])
			// fmt.Println("33333333333333333333333333")
			b.mu.Unlock()
		case unsubStream := <-b.unsubStream:
			// subscriptions are always 1:1, never 1:many, so we can
			// iterate through entire map to find our sub and delete
			// it
			// TODO: If we dont find a sub, return an error
			b.mu.Lock()
			for k, v := range b.subs {
				for idx, stream := range v[:] {
					if stream == unsubStream {
						b.subs[k] = removeIndex(b.subs[k], idx)
						// fmt.Println("11111111111111111111111111")
						// fmt.Printf("Closing stream: %v\n", stream)
						// fmt.Println("11111111111111111111111111")
						close(stream)
					}
				}
			}
			b.mu.Unlock()
		case msg := <-b.pubStream:
			// TODO: Need to groom out dead subscriptions?
			topics, err := findTopics(msg)
			if err != nil {
				// TODO: Better way to do this
				fmt.Println("Could not correlate server response to a topic: ", topics, err)
			}
			// fmt.Println("00000000000000000000000000")
			// fmt.Printf("Topics are: %v\n", topics)
			// fmt.Println("00000000000000000000000000")

			for _, topic := range topics {
				b.mu.Lock()
				subs := b.subs[topic][:]
				b.mu.Unlock()
				// fmt.Println("00000000000000000000000000")
				// fmt.Printf("Streams for topic %v are: %v\n", topic, b.subs[topic])
				// fmt.Println("00000000000000000000000000")
				for _, stream := range subs {
					// fmt.Printf("Publishing %v to topic %v\n", msg, topic)
					// TODO: Possible the listener is dead/gone, need to timeout?
					//       Or possibly use buffered channel with timeout?
					select {
					case <-b.done:
					case stream <- msg:
					}
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
	case stepOk:
		return []Topic{StepOkTopic}, nil
	default:
		// TODO: Rethink how I'm doing this
		fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		fmt.Printf("Unknown message type: %s\n", msg)
		fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		return []Topic{}, ErrUnknownMessageType
	}
}

func (b *Broker) Publish(msg interface{}) {
	b.pubStream <- msg
}

type topicSubscription struct {
	topic     Topic
	subStream chan interface{}
}

func (b *Broker) Subscribe(topic Topic) <-chan interface{} {
	subStream := make(chan interface{}, 1)
	t := topicSubscription{
		topic:     topic,
		subStream: subStream,
	}
	b.subStream <- t

	return subStream
}

func (b *Broker) Unsubscribe(unsub <-chan interface{}) {
	// TODO: How can I notify the caller that this failed?
	//       Maybe create and return an err channel, and pass
	//       err channel to Start?
	b.unsubStream <- unsub
}

func (b *Broker) SubscriptionCount() int {
	// TODO: Instead of returning value, requiring mutex locks,
	//	     return channel, and have start goroutine return the
	//		 current subscription count
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
