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
}
