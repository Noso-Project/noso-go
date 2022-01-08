package common

import (
	"reflect"
	"testing"
	"time"
)

func TestHashRateReport(t *testing.T) {
	t.Run("new HashRateReport", func(t *testing.T) {
		name := "TestMinerName"
		hashes := 1_000_000
		start := time.Now()
		stop := start.Add(10 * time.Second)
		got := NewHashRateReport(name, hashes, start, stop)
		want := HashRateReport{
			MinerName: name,
			HashRate:  100_000,
			hashes:    hashes,
			start:     start,
			stop:      stop,
			duration:  10 * time.Second,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("format string", func(t *testing.T) {
		name := "TestMinerName"
		hashes := 1_000_000
		start := time.Now()
		stop := start.Add(10 * time.Second)
		got := NewHashRateReport(name, hashes, start, stop).String()
		want := "100.000 Khash/s"

		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}
