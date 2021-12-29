package common

import (
	"reflect"
	"testing"
	"time"
)

func TestBroker(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		done := make(chan struct{}, 0)
		defer close(done)
		event, _ := parse(JOINOK_default)
		broker := NewBroker(done)
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
		done := make(chan struct{}, 0)
		defer close(done)
		broker := NewBroker(done)

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
		done := make(chan struct{}, 0)
		defer close(done)
		broker := NewBroker(done)
		ch := broker.Subscribe(JoinTopic)
		broker.Unsubscribe(ch)

		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Failed to close unsubscribed channel")
		}

		got := broker.SubscriptionCount()
		want := 0

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
	// t.Run("subscriber not listening", func(t *testing.T) {
	// 	done := make(chan struct{}, 0)
	// 	defer close(done)
	// 	event, _ := parse(PONG_default)
	// 	broker := NewBroker(done)
	// 	subCh := broker.Subscribe(PingPongTopic)
	// 	broker.Publish(event)

	// 	var got interface{}
	// 	select {
	// 	case got = <-subCh:
	// 	case <-time.After(100 * time.Millisecond):
	// 		t.Fatal("Failed to get message from broker")
	// 	}

	// 	want := event

	// 	if !reflect.DeepEqual(got, want) {
	// 		t.Errorf("got %v, want %v", got, want)
	// 	}
	// })
	t.Run("broker closes all subs on exit", func(t *testing.T) {
		done := make(chan struct{}, 0)
		event, _ := parse(PONG_default)
		broker := NewBroker(done)
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

		close(done)

		// TODO: refactor to helper function? assertChanClosed?
		select {
		case got := <-pingStream:
			if got != nil {
				t.Fatalf("got %v, want nil", got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for pingStream to close")
		}

		select {
		case got := <-joinStream:
			if got != nil {
				t.Fatalf("got %v, want nil", got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for joinStream to close")
		}

		select {
		case got := <-poolDataStream:
			if got != nil {
				t.Fatalf("got %v, want nil", got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timed out waiting for poolDataStream to close")
		}
	})
}
