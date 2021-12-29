package common

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestBroker(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		event, _ := parse(JOINOK_default)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		broker := NewBroker(ctx, cancel)
		subCh := broker.Subscribe(JoinTopic)
		broker.Publish(event)

		var got interface{}
		select {
		case got = <-subCh:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Failed to get message from broker")
		}

		want := event

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("subscribe", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		broker := NewBroker(ctx, cancel)

		got := broker.SubscriptionCount()
		want := 0

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}

		subCh := broker.Subscribe(JoinTopic)
		event, _ := parse(JOINOK_default)
		broker.Publish(event)
		<-subCh

		got = broker.SubscriptionCount()
		want = 1

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
	t.Run("unsubscribe", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		broker := NewBroker(ctx, cancel)
		stream := broker.Subscribe(JoinTopic)
		broker.Unsubscribe(stream)

		assertRoStreamClosed(t, stream)

		got := broker.SubscriptionCount()
		want := 0

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
	t.Run("broker closes all subs on exit", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		event, _ := parse(PONG_default)
		broker := NewBroker(ctx, cancel)
		pingStream := broker.Subscribe(PingPongTopic)
		joinStream := broker.Subscribe(JoinTopic)
		poolDataStream := broker.Subscribe(PoolDataTopic)

		// Verify the subscriptions are alive
		broker.Publish(event)

		pingStreamCp := pingStream
		poolDataStreamCp := poolDataStream
		for x := 0; x < 2; x++ {
			// a nil channel blocks indefinitely, making sure both
			// streams post exactly once
			select {
			case <-pingStreamCp:
				pingStreamCp = nil
			case <-poolDataStreamCp:
				poolDataStreamCp = nil
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timed out waiting for all subscriptions to get event")
			}
		}

		cancel()

		assertRoStreamClosed(t, pingStream)
		assertRoStreamClosed(t, joinStream)
		assertRoStreamClosed(t, poolDataStream)
	})
	t.Run("broker hung on publish", func(t *testing.T) {
		oldTimeout := PublishTimeout
		PublishTimeout = 1 * time.Millisecond
		defer func() { PublishTimeout = oldTimeout }()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		event, _ := parse(PONG_default)
		broker := NewBroker(ctx, cancel)
		pingStream := broker.Subscribe(PingPongTopic)
		broker.Subscribe(PingPongTopic)

		go func() {
			// Publish an event to trigger publish workflow
			broker.Publish(event)

			// Now send another event to ensure broker should be hung on write
			// to stream causing it to restart the client
			broker.Publish(event)
		}()

		select {
		case got := <-pingStream:
			if got == nil {
				t.Fatalf("got nil, want %v", got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Failed to get message from broker")
		}

		select {
		case <-broker.Done():
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for broker to call cancel func")
		}
	})
}

func assertRoStreamClosed(t *testing.T, stream <-chan interface{}) {
	t.Helper()

	select {
	case got := <-stream:
		if got != nil {
			t.Errorf("got %v, want nil", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timed out waiting for %v to close", stream)
	}
}
