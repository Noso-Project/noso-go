package miner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

const (
	hashChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func Miner(workerNum string, comms *Comms, ready chan bool) {
	var (
		jobStart    time.Time
		jobDuration time.Duration
		buff        *bytes.Buffer
		hashStr     string
		targets     []string
		targetLen   int
		targetMin   int

		// From hash_22
		seedLen   int
		w         rune
		x         rune
		y         rune
		z         rune
		tmp       [32]byte
		val       string
		hashCount int
	)

	encoded := make([]byte, 64)

	// Wait until ready
	<-ready

	for job := range comms.Jobs {
		jobStart = time.Now()
		targetMin = (job.Diff / 10) + 1 - job.PoolDepth
		buff = bytes.NewBuffer(job.SeedFullBytes)
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
						hashCount++
						buff.Truncate(seedLen)

						buff.WriteRune(w)
						buff.WriteRune(x)
						buff.WriteRune(y)
						buff.WriteRune(z)

						// This is the meat of the hashing
						tmp = sha256.Sum256(buff.Bytes())
						hex.Encode(encoded, tmp[:])
						val = BytesToString(encoded)

						// TODO: We could almost certainly increase hashrate if we
						//       could search the sha sum bytes rather than converting
						//       to a string first and then doing a string search
						// TODO: Benchmark doing a small substring search
						if !strings.Contains(val, targets[0]) {
							// targets[0] is that absolute minimum that a pool will accept
							// if we dont match that minimum, we can drop this solution
							// and continue with the hashing
							continue
						}

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

						comms.Solutions <- Solution{
							Seed:       job.SeedMiner,
							HashStr:    job.SeedPostfix + hashStr,
							Block:      job.Block,
							Chars:      job.TargetChars,
							Step:       job.Step,
							SolvedHash: *(*string)(unsafe.Pointer(&solution)),
							TargetLen:  targetLen,
							Target:     job.TargetString[:targetLen],
							FullTarget: job.TargetString[:job.TargetChars],
						}
					}
				}
			}
		}
		jobDuration = time.Since(jobStart)
		comms.Reports <- Report{WorkerNum: workerNum, Hashes: hashCount, Duration: jobDuration}
	}
}

func BytesToString(bytes []byte) string {
	var s string
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	stringHeader.Data = sliceHeader.Data
	stringHeader.Len = sliceHeader.Len
	return s
}
