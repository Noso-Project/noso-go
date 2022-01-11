package miner

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Noso-Project/noso-go/internal/common"
	simd "github.com/minio/sha256-simd"
)

const hashChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func MinerManager(ctx context.Context, client *common.Client, broker *common.Broker, opts common.Opts, wg *sync.WaitGroup) {
	// TODO: Investigate why this is needed
	time.Sleep(time.Second)
	wg.Done()
	jobStream := requestJobStream(ctx, broker)

	for x := 0; x < opts.Cpu; x++ {
		go func(name int) {
			var (
				job       common.Job
				targets   []string
				target    string
				targetMin int
				val       string
				jobStart  time.Time
				hashStr   string
				targetLen int
				fullHash  []byte
				hashed    [32]byte
				hashCount int
				i         int
			)

			encoded := make([]byte, 64)

			for job = range jobStream {
				jobStart = time.Now()
				targetMin = (job.Diff / 10) + 1 - job.PoolDepth
				seed := []byte(job.MinerSeed)
				fullHash = make([]byte, len(seed)+4)
				hashCount = 0

				targets = make([]string, job.PoolDepth+1)

				for i := 0; i < job.PoolDepth+1; i++ {
					targets[i] = job.TargetString[:targetMin+i]
				}

				// TODO: Update this comment, and make size adjustable at CLI
				// 5 was chosen so that it would take roughly 1 second to iterate
				// through all the hashes on one modern-ish cpu thread
				for _, h := range common.AllHashes() {
					hashCount++

					i = 0
					i += copy(fullHash, seed)
					copy(fullHash[i:], h)

					// This is the meat of the hashing
					// tmp = simd.Sum256(buff.Bytes())
					hashed = simd.Sum256(fullHash)
					hex.Encode(encoded, hashed[:])
					val = common.BytesToString(encoded)

					// TODO: We could almost certainly increase hashrate if we
					//       could search the sha sum bytes rather than converting
					//       to a string first and then doing a string search
					// TODO: Benchmark doing a small substring search
					// if !strings.Contains(val, targets[0]) {
					// fmt.Println("Seed is:      ", seed)
					// fmt.Println("Extr is:      ", h)
					// fmt.Println("fullHash is:  ", fullHash)
					// fmt.Println("target[0] is: ", targets[0])
					// fmt.Println("val is:       ", val)
					// time.Sleep(time.Second)
					if !strings.Contains(val, targets[0]) {
						// targets[0] is that absolute minimum that a pool will accept
						// if we dont match that minimum, we can drop this solution
						// and continue with the hashing
						continue
					}

					targetLen = targetMin
					for _, target = range targets[1:] {
						if !strings.Contains(val, target) {
							break
						}
						targetLen++
					}

					hashStr = string(h)
					solution := make([]byte, len(val))
					copy(solution, val)
					sol := fmt.Sprintf("%s", solution)

					fmt.Printf("Miner %2d found a %2d char solution - %s\n", name, targetLen, sol)
					go broker.Publish(ctx, common.Solution{
						Block:     job.Block,
						Seed:      job.MinerSeedBase,
						HashStr:   job.MinerPostfix + hashStr,
						TargetLen: targetLen,
					})
				}
				stop := time.Now()
				report := common.NewHashRateReport(fmt.Sprintf("Miner %2d", name), hashCount, jobStart, stop)
				// fmt.Printf("Miner %d - %s\n", name, report)
				go broker.Publish(ctx, report)

			}
		}(x)
	}

	select {
	case <-ctx.Done():
	}
}

func JoinSize(size int, s ...[]byte) []byte {
	b, i := make([]byte, size), 0
	for _, v := range s {
		i += copy(b[i:], v)
	}
	return b
}
