package miner

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	JOINOK     = "JOINOK"
	PASSFAILED = "PASSFAILED"
	PAYMENTOK  = "PAYMENTOK"
	PONG       = "PONG"
	POOLSTEPS  = "POOLSTEPS"
	STEPOK     = "STEPOK"
)

func Parse(comms *Comms, resp string) {
	if resp == "" {
		fmt.Println("Got an empty response")
		return
	}
	r := strings.Split(resp, " ")

	switch r[0] {
	case JOINOK:
		comms.PoolAddr <- r[1]
		comms.MinerSeed <- r[2]
		poolData(comms, r, 2)
		comms.Joined <- struct{}{}
	case PASSFAILED:
		fmt.Println("Incorrect pool password")
	case PAYMENTOK:
	case PONG:
		// NoOp
	case POOLSTEPS:
		poolData(comms, r, 0)
	case STEPOK:
		fmt.Println("Step solution accepted by pool")
		comms.StepSolved <- 1
	default:
		fmt.Printf("Uknown response code: %s\n", r[0])
	}
}

func poolData(comms *Comms, resp []string, offset int) {
	block, err := strconv.Atoi(resp[2+offset])
	if err != nil {
		fmt.Printf("Error converting target block: %s\n", resp[2+offset])
	} else {
		comms.Block <- block
	}

	comms.TargetString <- resp[3+offset]

	targetChars, err := strconv.Atoi(resp[4+offset])
	if err != nil {
		fmt.Printf("Error converting target chars: %s\n", resp[4+offset])
	} else {
		comms.TargetChars <- targetChars
	}

	step, err := strconv.Atoi(resp[5+offset])
	if err != nil {
		fmt.Printf("Error converting target chars: %s\n", resp[5+offset])
	} else {
		comms.Step <- step
	}

	diff, err := strconv.Atoi(resp[6+offset])
	if err != nil {
		fmt.Printf("Error converting target chars: %s\n", resp[6+offset])
	} else {
		comms.Diff <- diff
	}

	comms.Balance <- resp[7+offset]

	blocksTillPayment, err := strconv.Atoi(resp[8+offset])
	if err != nil {
		fmt.Printf("Error converting target chars: %s\n", resp[8+offset])
	} else {
		comms.BlocksTillPayment <- blocksTillPayment
	}
}
