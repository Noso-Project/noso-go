package miner

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Noso-Project/noso-go/internal/common"
	simd "github.com/minio/sha256-simd"
)

const hashChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func MinerManagerNew(ctx context.Context, client *common.Client, broker *common.Broker, wg *sync.WaitGroup) {
	time.Sleep(time.Second)
	wg.Done()
	jobStream := requestJobStream(ctx, broker)

	for x := 0; x < 12; x++ {
		go func(name int) {
			var (
				// hasher    *common.MultiStep256
				job       common.Job
				targets   []string
				targetMin int
				val       string
				// valBytes  []byte
				// solution string

				// jobStart time.Time
				// jobDuration time.Duration
				buff    *bytes.Buffer
				hashStr string
				// targets   []string
				targetLen int
				// targetMin int

				// // From hash_22
				seedLen int
				w       rune
				x       rune
				y       rune
				z       rune
				tmp     [32]byte
				// val     string
				hashCount int
			)

			encoded := make([]byte, 64)

			for job = range jobStream {
				// jobStart = time.Now()
				// 	hashCount := 0
				// 	// fmt.Printf("Pulled job: %+v\n", job)
				// 	hasher = common.NewMultiStep256(job.MinerSeed)

				// 	targetMin = (job.Diff / 10) + 1 - job.PoolDepth
				// 	targets = make([]string, job.PoolDepth+1)

				// 	for i := 0; i < job.PoolDepth+1; i++ {
				// 		targets[i] = job.TargetString[:targetMin+i]
				// 	}
				// 	// fmt.Printf("Targets are: %v\n", targets)

				// 	// fmt.Printf("Targets: %v\n", targets)

				// 	for valBytes = range job.GenBytes(ctx) {
				// 		// fmt.Println("Gen'd bytes:", valBytes)
				// 		// start := time.Now()
				// 		hashCount++

				// 		// hasher.Reset()
				// 		h := hasher.HashTest(valBytes)
				// 		solution = hasher.Search(targets)
				// 		// fmt.Printf("Took %s to do hash and first search\n", time.Since(start))
				// 		// time.Sleep(time.Second)

				// 		if solution != "" {
				// 			fmt.Printf("Miner %2d found a %2d char solution - %v\n", name, len(solution), h)
				// 			broker.Publish(ctx, common.Solution{
				// 				Block:     job.Block,
				// 				Seed:      job.MinerSeedBase,
				// 				HashStr:   job.MinerPostfix + val,
				// 				TargetLen: len(solution),
				// 			})
				// 		}
				// 	}
				// 	stop := time.Now()
				// 	report := common.NewHashRateReport(fmt.Sprintf("Miner %2d", name), 2_383_280, jobStart, stop)
				// 	fmt.Printf("%s - %s\n", report.MinerName, report)
				// }
				// jobStart := time.Now()
				targetMin = (job.Diff / 10) + 1 - job.PoolDepth
				buff = bytes.NewBufferString(job.MinerSeed)
				seedLen = buff.Len()
				hashCount = 0

				targets = make([]string, job.PoolDepth+1)

				for i := 0; i < job.PoolDepth+1; i++ {
					targets[i] = job.TargetString[:targetMin+i]
				}

				// 5 was chosen so that it would take roughly 1 second to iterate
				// through all the hashes on one modern-ish cpu thread
				for _, w = range hashChars[:5] {
					for _, x = range hashChars {
						for _, y = range hashChars {
							for _, z = range hashChars {
								// start := time.Now()
								hashCount++
								buff.Truncate(seedLen)

								buff.WriteRune(w)
								buff.WriteRune(x)
								buff.WriteRune(y)
								buff.WriteRune(z)

								// This is the meat of the hashing
								tmp = simd.Sum256(buff.Bytes())
								hex.Encode(encoded, tmp[:])
								val = common.BytesToString(encoded)

								// TODO: We could almost certainly increase hashrate if we
								//       could search the sha sum bytes rather than converting
								//       to a string first and then doing a string search
								// TODO: Benchmark doing a small substring search
								if !strings.Contains(val, targets[0]) {
									// targets[0] is that absolute minimum that a pool will accept
									// if we dont match that minimum, we can drop this solution
									// and continue with the hashing
									// fmt.Printf("Took %s to do hash and first search\n", time.Since(start))
									// time.Sleep(time.Second)
									continue
								}
								// time.Sleep(time.Second)

								targetLen = targetMin
								for _, t := range targets[1:] {
									if !strings.Contains(val, t) {
										break
									}
									targetLen++
								}

								hashStr = string(w) + string(x) + string(y) + string(z)
								solution := make([]byte, len(val))
								copy(solution, val)
								sol := fmt.Sprintf("%s", solution)

								fmt.Printf("Miner %2d found a %2d char solution - %s\n", name, targetLen, sol)
								broker.Publish(ctx, common.Solution{
									Block:     job.Block,
									Seed:      job.MinerSeedBase,
									HashStr:   job.MinerPostfix + hashStr,
									TargetLen: targetLen,
								})
							}
						}
					}
				}
				// stop := time.Now()
				// report := common.NewHashRateReport(fmt.Sprintf("Miner %2d", name), 2_383_280, jobStart, stop)
				// fmt.Printf("%s - %s\n", report.MinerName, report)
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
