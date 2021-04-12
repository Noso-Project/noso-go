package miner

import (
	"strings"
)

func JobFeeder(comms *Comms) {
	var (
		poolAddr     string
		minerSeed    string
		targetChars  int
		targetBlock  int
		targetString string
		currentStep  int
		job          Job
		running      bool
	)

	for poolAddr == "" || minerSeed == "" {
		select {
		case poolAddr = <-comms.NewPoolAddr:
		case minerSeed = <-comms.NewMinerSeed:
		}
	}

	for {
		for x := 0; x < 92; x++ {
			for y := 0; y < 92; y++ {
				for z := 0; z < 92; z++ {
					seedBase := minerSeed[:len(minerSeed)-3]
					seedX := string(rune('!' + x))
					seedY := string(rune('!' + y))
					seedZ := string(rune('!' + z))
					seed := seedBase + seedX + seedY + seedZ

					for num := 1; num < 999; num++ {
						running = true
						for running {
							job = Job{
								Start:        num * 10000000,
								Stop:         (num * 10000000) + 9999999,
								TargetChars:  targetChars,
								TargetBlock:  targetBlock,
								TargetString: strings.ToLower(targetString),
								Seed:         seed,
								PoolAddr:     poolAddr,
								Step:         currentStep,
							}
							select {
							case targetChars = <-comms.NewChars:
								job.TargetChars = targetChars
							case targetBlock = <-comms.NewBlock:
								job.TargetBlock = targetBlock
							case targetString = <-comms.NewString:
								job.TargetString = targetString
							case currentStep = <-comms.NewStep:
								job.Step = currentStep
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
