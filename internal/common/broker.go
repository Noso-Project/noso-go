package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
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
	case SolutionTopic:
		return "SolutionTopic"
	case JobTopic:
		return "JobTopic"
	default:
		return fmt.Sprintf("%d (cant find string)", int(t))
	}
}

const (
	JobTopic Topic = iota + 1
	JoinTopic
	PingPongTopic
	PoolDataTopic
	PoolStepsTopic
	SolutionTopic
	StepOkTopic
)

var (
	PublishTimeout        = 1 * time.Second
	SubscribeTimeout      = 1 * time.Second
	ErrSubscribeTimeout   = errors.New("Timed out trying to subscribe to Topic")
	ErrUnknownMessageType = errors.New("Could not correlate server response to topic")
)

// TODO: find a way to not use interface{} for the channel
//       - Already boned on this once
type Broker struct {
	ctx            context.Context
	cancel         context.CancelFunc
	pubStream      chan interface{}
	subStream      chan topicSubscription
	unsubStream    chan (<-chan interface{})
	subCountReq    chan chan int
	publishTimeout time.Duration
}

func NewBroker(ctx context.Context, cancel context.CancelFunc) (b *Broker) {
	InitLogger(os.Stdout)
	b = new(Broker)
	b.ctx = ctx
	b.cancel = cancel
	b.pubStream = make(chan interface{}, 0)
	b.subStream = make(chan topicSubscription, 0)
	b.unsubStream = make(chan (<-chan interface{}), 0)
	b.subCountReq = make(chan chan int, 0)
	b.publishTimeout = PublishTimeout

	var wg sync.WaitGroup
	wg.Add(1)
	go b.start(&wg)
	wg.Wait()

	return
}

func removeIndex(s []chan interface{}, index int) []chan interface{} {
	return append(s[:index], s[index+1:]...)
}

func (b *Broker) Done() <-chan struct{} {
	return b.ctx.Done()
}

func (b *Broker) start(wg *sync.WaitGroup) {
	wg.Done()
	subs := make(map[Topic][]chan interface{})
	subs[JoinTopic] = make([]chan interface{}, 0)
	subs[PingPongTopic] = make([]chan interface{}, 0)
	for {
		select {
		case <-b.ctx.Done():
			logger.Debug("Entering <-b.ctx.Done()")
			// Attempt to close every stream in subs
			// TODO: Make this safer, so closing a closed stream doesnt panic
			for _, v := range subs {
				for _, stream := range v[:] {
					close(stream)
				}
			}
			logger.Debug("Leaving <-b.ctx.Done()")
			return
		case sub := <-b.subStream:
			logger.Debug("Entering sub := <-b.subStream")
			logger.Debugf("About to sub: topic %v channel %v", sub.topic, sub.subStream)
			subs[sub.topic] = append(subs[sub.topic], sub.subStream)
			logger.Debugf("Topics subs now: %v", subs[sub.topic])
			// TODO: Return an err here, if there is one to return?
			sub.errStream <- nil
			logger.Debug("Leaving sub := <-b.subStream")
		case unsubStream := <-b.unsubStream:
			logger.Debug("Entering unsubStream := <-b.unsubStream")
			// subscriptions are always 1:1, never 1:many, so we can
			// iterate through entire map to find our sub and delete
			// it
			// TODO: If we dont find a sub, return an error
			for k, v := range subs {
				for idx, stream := range v[:] {
					if stream == unsubStream {
						subs[k] = removeIndex(subs[k], idx)
						logger.Debugf("Closing stream: %v", stream)
						close(stream)
					}
				}
			}
			logger.Debug("Leaving unsubStream := <-b.unsubStream")
		case msg := <-b.pubStream:
			logger.Debug("Entering msg := <-b.pubStream")
			topics, err := findTopics(msg)
			if err != nil {
				// TODO: Better way to do this
				logger.Debug("Could not correlate server response to a topic: ", topics, err)
			}
			logger.Debugf("Topics are: %v", topics)

		loop:
			for _, topic := range topics {
				logger.Debugf("Streams for topic %v are: %v", topic, subs[topic])
				for _, stream := range subs[topic] {
					logger.Debugf("Publishing %v to %v stream for %v", msg, stream, topic)
					select {
					case <-b.ctx.Done():
					case stream <- msg:
					case <-time.After(b.publishTimeout):
						// TODO: Need to log this as an error visible to user
						logger.Debugf("Client broker is hung on write to %v stream for %s topic", stream, topic)
						b.cancel()
						break loop
					}
				}
				logger.Debug("msg := <-b.pubStream released the lock")
			}
			logger.Debug("Leaving msg := <-b.pubStream")
		case subCountStream := <-b.subCountReq:
			subCount := 0
			for _, v := range subs {
				subCount += len(v)
			}
			select {
			case <-b.Done():
				return
			case subCountStream <- subCount:
			}
		}
	}
}

func findTopics(msg interface{}) ([]Topic, error) {
	switch msg.(type) {
	case JoinOk:
		return []Topic{JoinTopic, PoolDataTopic}, nil
	case PassFailed:
		return []Topic{JoinTopic}, nil
	case AlreadyConnected:
		return []Topic{JoinTopic}, nil
	case Pong:
		return []Topic{PingPongTopic, PoolDataTopic}, nil
	case PoolSteps:
		return []Topic{PoolStepsTopic, PoolDataTopic}, nil
	case StepOk:
		return []Topic{StepOkTopic}, nil
	case Solution:
		return []Topic{SolutionTopic}, nil
	case Job:
		return []Topic{JobTopic}, nil
	case JobStreamReq:
		return []Topic{JobTopic}, nil
	default:
		// TODO: Rethink how I'm doing this
		logger.Debug("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		logger.Debugf("Unknown message type: %s", msg)
		logger.Debug("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
		return []Topic{}, ErrUnknownMessageType
	}
}

func (b *Broker) Publish(msg interface{}) {
	b.pubStream <- msg
}

type topicSubscription struct {
	topic     Topic
	subStream chan interface{}
	errStream chan error
}

func (b *Broker) Subscribe(topic Topic) (<-chan interface{}, error) {
	// TODO: Confusing naming between this and b.subStream, rethink
	subStream := make(chan interface{}, 0)
	errStream := make(chan error, 0)
	defer close(errStream)
	t := topicSubscription{
		topic:     topic,
		subStream: subStream,
		errStream: errStream,
	}
	b.subStream <- t

	select {
	case err := <-errStream:
		return subStream, err
	case <-time.After(time.Second):
		return subStream, ErrSubscribeTimeout
	}
}

func (b *Broker) Unsubscribe(unsub <-chan interface{}) {
	// TODO: How can I notify the caller that this failed?
	//       Maybe create and return an err channel, and pass
	//       err channel to Start?
	logger.Debug("Entering Broker Unsubscribe")
	b.unsubStream <- unsub
	logger.Debug("Leaving Broker Unsubscribe")
}

func (b *Broker) SubscriptionCount() int {
	logger.Debug("Entering SubscriptionCount")
	subCountStream := make(chan int, 0)
	defer close(subCountStream)
	b.subCountReq <- subCountStream
	select {
	case <-b.Done():
		return -1
	case count := <-subCountStream:
		logger.Debugf("Received %d subscriptions count from subCountStream", count)
		return count
	}
}
