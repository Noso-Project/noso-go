package miner

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	JOINOK    = "JOINOK"
	PONG      = "PONG"
	POOLSTEPS = "POOLSTEPS"
)

func Parse(comms Comms, resp string) {
	r := strings.Split(resp, " ")

	switch r[0] {
	case JOINOK:
		comms.PoolAddr <- r[1]
		comms.MinerSeed <- r[2]
		poolData(comms, r, 2)
	case POOLSTEPS:
		poolData(comms, r, 0)
	case PONG:
	default:
		fmt.Printf("Uknown response code: %s\n", r[0])
	}
}

func poolData(comms Comms, resp []string, offset int) {
	targetBlock, err := strconv.Atoi(resp[2+offset])
	if err != nil {
		fmt.Printf("Error converting target block: %s\n", resp[2+offset])
	} else {
		comms.TargetBlock <- targetBlock
	}

	comms.TargetString <- resp[3+offset]

	targetChars, err := strconv.Atoi(resp[4+offset])
	if err != nil {
		fmt.Printf("Error converting target chars: %s\n", resp[4+offset])
	} else {
		comms.TargetChars <- targetChars
	}

	currentStep, err := strconv.Atoi(resp[5+offset])
	if err != nil {
		fmt.Printf("Error converting target chars: %s\n", resp[5+offset])
	} else {
		comms.CurrentStep <- currentStep
	}
}
