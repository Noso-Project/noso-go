package miner

import (
	"strings"
)

func NewJobComms() *JobComms {
	return &JobComms{
		PoolAddr:     make(chan string, 10),
		MinerSeed:    make(chan string, 10),
		Block:        make(chan int, 10),
		Step:         make(chan int, 10),
		Diff:         make(chan int, 10),
		TargetChars:  make(chan int, 10),
		TargetString: make(chan string, 10),
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
	PoolAddr     string
	Seed         string
	TargetString string
	TargetChars  int
	Diff         int
	Block        int
	Step         int
	Start, Stop  int
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
	)

	// Step is the only int that can actually be 0
	step = -1

	ready := make(chan struct{}, 10)

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

	for {
		for x := 0; x < 92; x++ {
			for y := 0; y < 92; y++ {
				for z := 0; z < 92; z++ {
					seedBase := minerSeed[:len(minerSeed)-3]
					seedX := string(rune('!' + x))
					seedY := string(rune('!' + y))
					seedZ := string(rune('!' + z))
					seed := seedBase + seedX + seedY + seedZ

					// "_" and "(" are reserved characters in Noso
					if strings.Contains(seed, "_") || strings.Contains(seed, "(") {
						continue
					}

					for num := 1; num < 9999; num++ {
						running = true
						for running {
							job = Job{
								Start:        num * 1000000,
								Stop:         (num * 1000000) + 999999,
								TargetString: strings.ToLower(targetString),
								TargetChars:  targetChars,
								Diff:         diff,
								Block:        block,
								Seed:         seed,
								PoolAddr:     poolAddr,
								Step:         step,
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
