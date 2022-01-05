package miner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Noso-Project/noso-go/internal/common"
)

const (
	JOINOK    = "JOINOK N6VxgLSpbni8kLbyUAjYXdHCPt2VEp 020000000 PoolData 37873 E1151A4F79E6394F6897A913ADCD476B 11 0 102 0 -30 42270 3"
	PONG      = "PONG PoolData 37892 C74B9ABA60E2EE1B52613959D4F06876 11 0 105 0 -29 86070 3"
	POOLSTEPS = "POOLSTEPS PoolData %d AD23A982B87D193E8384EB50C3F0B50C 11 %d 106 0 -23 43328 3"
)

func TestJobManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	conn, svr := net.Pipe()
	defer conn.Close()
	defer svr.Close()

	client := common.NewClientWithConn(ctx, conn)

	wg.Add(1)
	go JobManager(ctx, client, &wg)
	wg.Wait()

	// A miner will publish a request to the JobTopic, requesting a jobStream,
	// and including in the request a channel to receive the jobStream on
	jobStream := requestJobStream(ctx, client)

	// Send a JOINOK from svr to client so PoolData info gets published
	// The JobManager will get the PoolData message, build a new Job,
	// and put it on the jobStream
	fmt.Fprintln(svr, JOINOK)

	var job common.Job
	select {
	case job = <-jobStream:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for job from jobStream")
	}

	assertMsgAttrs(t, job.PoolAddr, "N6VxgLSpbni8kLbyUAjYXdHCPt2VEp")
	assertMsgAttrs(t, job.TargetString, "E1151A4F79E6394F6897A913ADCD476B")
	assertMsgAttrs(t, job.TargetChars, 11)
	assertMsgAttrs(t, job.Diff, 102)
	assertMsgAttrs(t, job.Block, 37873)
	assertMsgAttrs(t, job.Step, 0)
	assertMsgAttrs(t, job.PoolDepth, 3)
}

func assertMsgAttrs(t *testing.T, got, want interface{}) {
	t.Helper()

	switch got.(type) {
	case string:
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	case int:
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	}

}

func BenchmarkJobManager(b *testing.B) {
	b.Logf("b.N is: %d\n", b.N)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	conn, svr := net.Pipe()
	defer conn.Close()
	defer svr.Close()

	client := common.NewClientWithConn(ctx, conn)

	wg.Add(1)
	go JobManager(ctx, client, &wg)
	wg.Wait()

	// A miner will publish a request to the JobTopic, requesting a jobStream,
	// and including in the request a channel to receive the jobStream on
	jobStream := requestJobStream(ctx, client)

	// Send a JOINOK from svr to client so PoolData info gets published
	// The JobManager will get the PoolData message, build a new Job,
	// and put it on the jobStream
	go func() {
		fmt.Fprintln(svr, JOINOK)
		pongTicker := time.NewTicker(11 * time.Millisecond)
		poolStepsTicker := time.NewTicker(17 * time.Millisecond)
		for count := 0; ; count++ {
			select {
			case <-pongTicker.C:
				fmt.Fprintln(svr, PONG)
			case <-poolStepsTicker.C:
				fmt.Fprintln(svr, fmt.Sprintf(POOLSTEPS, count, count%10))
			}
		}
	}()

	for n := 0; n < b.N; n++ {
		<-jobStream
	}
}
