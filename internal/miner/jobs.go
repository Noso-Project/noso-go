package miner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func NewJobComms() *JobComms {
	return &JobComms{
		PoolAddr:     make(chan string, 0),
		MinerSeed:    make(chan string, 0),
		Block:        make(chan int, 0),
		Step:         make(chan int, 0),
		Diff:         make(chan int, 0),
		TargetChars:  make(chan int, 0),
		TargetString: make(chan string, 0),
	}
}

type JobComms struct {
	PoolAddr     chan string
	MinerSeed    chan string
	Block        chan int
	Step         chan int
	Diff         chan int
	TargetChars  chan int
	TargetString chan string
}

type Job struct {
	PoolAddr      string
	SeedMiner     string
	SeedPostfix   string
	SeedFull      string
	SeedFullBytes []byte
	TargetString  string
	TargetChars   int
	Diff          int
	Block         int
	Step          int
}

func JobFeeder(comms *Comms, jobComms *JobComms) {
	var (
		poolAddr     string
		minerSeed    string
		block        int
		step         int
		diff         int
		targetChars  int
		targetString string
		job          Job
		running      bool
		postfix      string
	)

	verSha := sha256.Sum256([]byte(MinerName))
	verShaHex := hex.EncodeToString(verSha[:])
	ver := verShaHex[:2]

	// Step is the only int that can actually be 0
	step = -1

	ready := make(chan struct{}, 0)

	// This is pretty ugly, there has to be a better way
	go func() {
		for {
			select {
			case poolAddr = <-jobComms.PoolAddr:
			case minerSeed = <-jobComms.MinerSeed:
			case targetChars = <-jobComms.TargetChars:
			case targetString = <-jobComms.TargetString:
			case diff = <-jobComms.Diff:
			case block = <-jobComms.Block:
			case step = <-jobComms.Step:
			}

			if poolAddr == "" {
				continue
			} else if minerSeed == "" {
				continue
			} else if block == 0 {
				continue
			} else if step == -1 {
				continue
			} else if diff == 0 {
				continue
			} else if targetChars == 0 {
				continue
			} else if targetString == "" {
				continue
			} else if minerSeed == "" {
				continue
			} else {
				ready <- struct{}{}
				break
			}
		}
	}()

	<-ready

	// ! is 33
	// % is 37
	// ( is 40
	// _ is 95
	// _ and ( are reserved chars, skip them
	// Wallet doesn't like %, skip it

	for {
		for x := 0; x < 92; x++ {
			if x == 4 || x == 7 || x == 62 {
				continue
			}
			for y := 0; y < 92; y++ {
				if y == 4 || y == 7 || y == 62 {
					continue
				}
				for z := 0; z < 92; z++ {
					if z == 4 || z == 7 || z == 62 {
						continue
					}
					seedBase := minerSeed[:len(minerSeed)-3]
					seedX := fmt.Sprint('!' + x)
					seedY := fmt.Sprint('!' + y)
					seedZ := fmt.Sprint('!' + z)
					seed := seedBase + seedX + seedY + seedZ

					// "_" and "(" are reserved characters in Noso
					if strings.Contains(seed, "_") || strings.Contains(seed, "(") {
						continue
					}

					for num := 1; num < 999; num++ {
						postfix = ver + fmt.Sprintf("%03d", num)
						running = true
						fullSeed := seed + poolAddr + postfix
						fullSeedBytes := []byte(fullSeed)
						for running {
							job = Job{
								TargetString:  strings.ToLower(targetString),
								TargetChars:   targetChars,
								Diff:          diff,
								Block:         block,
								SeedMiner:     seed,
								SeedPostfix:   postfix,
								SeedFull:      fullSeed,
								SeedFullBytes: fullSeedBytes,
								PoolAddr:      poolAddr,
								Step:          step,
							}
							select {
							case poolAddr = <-jobComms.PoolAddr:
							case minerSeed = <-jobComms.MinerSeed:
							case targetChars = <-jobComms.TargetChars:
								job.TargetChars = targetChars
							case targetString = <-jobComms.TargetString:
								job.TargetString = targetString
							case diff = <-jobComms.Diff:
								job.Diff = diff
							case block = <-jobComms.Block:
								job.Block = block
							case step = <-jobComms.Step:
								job.Step = step
							case comms.Jobs <- job:
								running = false
							}
						}
					}
				}
			}
		}
	}
}
