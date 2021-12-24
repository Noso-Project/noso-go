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
		event, _ := parse(JOINOK_1)
		broker := NewBroker(done)
		subCh := broker.Subscribe(JOINOK)
		broker.Publish(event)

		var got interface{}
		select {
		case got = <-subCh:
		case <-time.After(1000 * time.Millisecond):
			t.Fatal("Failed to get message from broker")
		}

		want := event

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
