package miner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

func Miner(worker_num string, comms *Comms, ready chan bool) {
	var (
		hash_bytes   [32]byte
		hash         string
		num          int
		target_len   int
		target_large string
		target_small string
	)

	// Wait until ready
	<-ready

	// Search for TargetChars - 1 solutions
	// Report any TargetChars solutions immediately
	// Store any TargetChars - 1 solutions until the steps drop
	for job := range comms.Jobs {
		target_large = job.TargetString[:job.TargetChars]
		target_small = job.TargetString[:job.TargetChars-1]
		for num = job.Start; num < job.Stop; num++ {
			hash_bytes = sha256.Sum256([]byte(job.Seed + job.PoolAddr + strconv.Itoa(num)))
			hash = hex.EncodeToString(hash_bytes[:])
			if !strings.Contains(hash, target_small) {
				continue
			} else if strings.Contains(hash, target_large) {
				target_len = job.TargetChars
			} else {
				target_len = job.TargetChars - 1
			}

			comms.Solutions <- Solution{
				Seed:        job.Seed,
				HashNum:     num,
				TargetBlock: job.TargetBlock,
				TargetChars: target_len,
			}

			fmt.Printf(
				found_one,
				worker_num,
				job.Block,
				job.Step,
				job.Seed,
				job.PoolAddr,
				num,
				hash,
				target_len,
				job.TargetString[:target_len],
				job.TargetString[:job.TargetChars],
			)
		}
		comms.Reports <- Report{WorkerNum: worker_num, Hashes: job.Stop - job.Start}
	}
}

const found_one = `
************************************
FOUND ONE
Worker        : %s
Block         : %d
Step          : %d
Seed          : %s
Pool Addr     : %s
Number        : %d
Found         : %s
Target Len    : %d
Target        : %s
Full Target   : %s
************************************
`
