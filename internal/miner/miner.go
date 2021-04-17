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

func Miner(worker_num string, comms *Comms, ready chan bool) {
	var (
		jobStart     time.Time
		jobDuration  time.Duration
		buff         *bytes.Buffer
		hashStr      string
		target_len   int
		target_large string
		target_small string

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

	// Search for TargetChars - 1 solutions
	// Report any TargetChars solutions immediately
	// Store any TargetChars - 1 solutions until the steps drop
	for job := range comms.Jobs {
		jobStart = time.Now()
		target_large = job.TargetString[:job.TargetChars]
		target_small = job.TargetString[:job.TargetChars-1]
		buff = bytes.NewBuffer(job.SeedFullBytes)
		seedLen = buff.Len()
		hashCount = 0
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

						// We could almost certainly increase hashrate if we
						// could search the sha sum bytes rather than converting
						// to a string first and then doing a string search
						// Also, need to benchmark doing a small substring search
						if !strings.Contains(val, target_small) {
							continue
						} else if strings.Contains(val, target_large) {
							target_len = job.TargetChars
						} else {
							target_len = job.TargetChars - 1
						}

						hashStr = string(w) + string(x) + string(y) + string(z)
						solution := make([]byte, len(val))
						copy(solution, val)

						comms.Solutions <- Solution{
							Seed:       job.SeedMiner,
							HashStr:    job.SeedPostfix + hashStr,
							Block:      job.Block,
							Chars:      target_len,
							Step:       job.Step,
							SolvedHash: *(*string)(unsafe.Pointer(&solution)),
							TargetLen:  target_len,
							Target:     job.TargetString[:target_len],
							FullTarget: job.TargetString[:job.TargetChars],
						}
					}
				}
			}
		}
		jobDuration = time.Since(jobStart)
		comms.Reports <- Report{WorkerNum: worker_num, Hashes: hashCount, Duration: jobDuration}
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
