package miner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ! is 33
// % is 37
// ( is 40
// _ is 95
// _ and ( are reserved chars, skip them
// Wallet doesn't like %, skip it
const (
	hashableSeedChars = "!\"#$&')*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^`abcdefghijklmnopqrstuvwxyz{|"
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
		PoolDepth:    make(chan int, 0),
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
	PoolDepth    chan int
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
	PoolDepth     int
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
		postfix      string
		poolDepth    int
	)

	verSha := sha256.Sum256([]byte(MinerName))
	verShaHex := hex.EncodeToString(verSha[:])
	ver := verShaHex[:2]

waitready:
	for {
		// When this channel is closed, it indicates a disconnected state
		disconnected := comms.Disconnected

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
				case poolDepth = <-jobComms.PoolDepth:
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
				} else if poolDepth == 0 {
					continue
				} else if minerSeed == "" {
					continue
				} else {
					close(ready)
					return
				}
			}
		}()

		<-ready

		// Randomize seed chars so that if a miner restarts in the middle of a block,
		// it isn't rehashing already hashed values
		seedChars := []rune(hashableSeedChars)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(seedChars), func(i, j int) { seedChars[i], seedChars[j] = seedChars[j], seedChars[i] })

		for {
			for _, x := range seedChars {
				for _, y := range seedChars {
					for _, z := range seedChars {

						seedBase := minerSeed[:len(minerSeed)-3]
						seedX := string(x)
						seedY := string(y)
						seedZ := string(z)
						seed := seedBase + seedX + seedY + seedZ

					loop:
						for num := 1; num < 999; num++ {
							postfix = ver + fmt.Sprintf("%03d", num)
							fullSeed := seed + poolAddr + postfix
							fullSeedBytes := []byte(fullSeed)
							for {
								job = Job{
									TargetString:  strings.ToLower(targetString),
									TargetChars:   targetChars,
									PoolDepth:     poolDepth,
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
								case poolDepth = <-jobComms.PoolDepth:
									job.PoolDepth = poolDepth
								case diff = <-jobComms.Diff:
									job.Diff = diff
								case block = <-jobComms.Block:
									job.Block = block
								case step = <-jobComms.Step:
									job.Step = step
								case comms.Jobs <- job:
									continue loop
								case <-disconnected:
									time.Sleep(10 * time.Second)
									continue waitready
								}
							}
						}
					}
				}
			}
		}
	}
}
