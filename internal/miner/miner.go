package miner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

const (
	hashChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func Miner(worker_num string, comms *Comms, ready chan bool) {
	var (
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

						tmp = sha256.Sum256(buff.Bytes())
						hex.Encode(encoded, tmp[:])
						val = BytesToString(encoded)

						if !strings.Contains(val, target_small) {
							continue
						} else if strings.Contains(val, target_large) {
							target_len = job.TargetChars
						} else {
							target_len = job.TargetChars - 1
						}

						hashStr = string(w) + string(x) + string(y) + string(z)

						comms.Solutions <- Solution{
							Seed:    job.SeedMiner,
							HashStr: job.SeedPostfix + hashStr,
							Block:   job.Block,
							Chars:   target_len,
							Step:    job.Step,
						}

						fmt.Printf(
							found_one,
							worker_num,
							job.Block,
							job.Step,
							job.SeedMiner,
							job.PoolAddr,
							job.SeedPostfix+hashStr,
							val,
							target_len,
							job.TargetString[:target_len],
							job.TargetString[:job.TargetChars],
						)
					}
				}
			}
		}
		comms.Reports <- Report{WorkerNum: worker_num, Hashes: hashCount}
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

const found_one = `
************************************
FOUND ONE
Worker        : %s
Block         : %d
Step          : %d
Seed          : %s
Pool Addr     : %s
Number        : %s
Found         : %s
Target Len    : %d
Target        : %s
Full Target   : %s
************************************
`
