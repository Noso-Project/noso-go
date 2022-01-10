package miner

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Noso-Project/noso-go/internal/common"
)

func MinerManagerNew(ctx context.Context, client *common.Client, broker *common.Broker, wg *sync.WaitGroup) {
	time.Sleep(time.Second)
	wg.Done()
	jobStream := requestJobStream(ctx, broker)

	for x := 0; x < 1; x++ {
		go func(name int) {
			var (
				hasher    *common.MultiStep256
				job       common.Job
				targets   []string
				targetMin int
				val       string
				valBytes  []byte
				solution  string
			)

			for job = range jobStream {
				jobStart := time.Now()
				hashCount := 0
				// fmt.Printf("Pulled job: %+v\n", job)
				hasher = common.NewMultiStep256(job.MinerSeed)

				targetMin = (job.Diff / 10) + 1 - job.PoolDepth
				targets = make([]string, job.PoolDepth+1)

				for i := 0; i < job.PoolDepth+1; i++ {
					targets[i] = job.TargetString[:targetMin+i]
				}
				// fmt.Printf("Targets are: %v\n", targets)

				// fmt.Printf("Targets: %v\n", targets)

				for valBytes = range job.GenBytes(ctx) {
					// fmt.Println("Gen'd bytes:", valBytes)
					// start := time.Now()
					hashCount++

					// hasher.Reset()
					h := hasher.HashTest(valBytes)
					solution = hasher.Search(targets)
					// fmt.Printf("Took %s to do hash and first search\n", time.Since(start))
					// time.Sleep(time.Second)

					if solution != "" {
						fmt.Printf("Miner %2d found a %2d char solution - %v\n", name, len(solution), h)
						broker.Publish(ctx, common.Solution{
							Block:     job.Block,
							Seed:      job.MinerSeedBase,
							HashStr:   job.MinerPostfix + val,
							TargetLen: len(solution),
						})
					}
				}
				stop := time.Now()
				report := common.NewHashRateReport(fmt.Sprintf("Miner %2d", name), 2_383_280, jobStart, stop)
				fmt.Printf("%s - %s\n", report.MinerName, report)
			}
		}(x)
	}

	select {
	case <-ctx.Done():
	}
}

func MinerManager(ctx context.Context, client *common.Client, broker *common.Broker, wg *sync.WaitGroup) {
	wg.Done()

	// TODO: some sort of race condition here, where we can request this stream too
	//       fast and then never get any jobs from the job manager
	time.Sleep(time.Second)
	rand.Seed(time.Now().UnixNano())
	mine := func(ctx context.Context, num int) {
		jobStream := requestJobStream(ctx, broker)
		// fmt.Printf("Miner %3d is about to start mining\n", num)

		d := time.Duration(rand.Intn(1000) + 500)
		dur := d * time.Millisecond
		ticker := time.NewTicker(dur)

		countInterval := rand.Intn(10) + 5

		var job, nilJob common.Job
		nilJob = common.Job{}
		job = nilJob
		// TODO: Implement orDone here, page 119
		for count := 0; ; count++ {
			// Change the interval every X loops
			if count%countInterval == 0 {
				countInterval = rand.Intn(10) + 5
				d = time.Duration(rand.Intn(1000) + 500)
				dur = d * time.Millisecond
				ticker.Reset(dur)

			}
			select {
			case <-job.Done():
				// fmt.Printf("Miner %5d job cancelled before completion", num)
				job = nilJob
			case <-ticker.C:
				select {
				case job = <-jobStream:
					// fmt.Printf("Miner %d got a new job: %+v\n", num, job)
					// fmt.Printf("Miner %5d got a new job\n", num)
				case <-time.After(1000 * time.Millisecond):
					// panic("Timed out waiting for a new job from job manager")
				case <-ctx.Done():
					return
				}
			case <-time.After(20000 * time.Millisecond):
				panic("Timed out waiting to request a new job from job manager")
			case <-ctx.Done():
				return
			}
		}
	}

	// for x := 0; x < 1; x++ {
	// for x := 0; x < 8192; x++ {
	for x := 0; x < 16384; x++ {
		// for x := 0; x < 32768; x++ {
		// for x := 0; x < 65536; x++ {
		// for x := 0; x < 131072; x++ {
		go mine(ctx, x)
	}

	select {
	case <-ctx.Done():
		return
	}
}
