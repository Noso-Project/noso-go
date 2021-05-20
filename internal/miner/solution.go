package miner

import "fmt"

func NewSolutionComms(sendChan chan string) *SolutionComms {
	return &SolutionComms{
		Block:    make(chan int, 0),
		Step:     make(chan int, 0),
		Diff:     make(chan int, 0),
		Solution: make(chan Solution, 0),
		SendChan: sendChan,
		StepSent: make(chan struct{}, 0),
	}
}

type SolutionComms struct {
	Block    chan int
	Step     chan int
	Diff     chan int
	Solution chan Solution
	SendChan chan string
	StepSent chan struct{}
}

type Solution struct {
	Seed       string
	HashStr    string
	Block      int
	Chars      int
	Step       int
	SolvedHash string
	TargetLen  int
	Target     string
	FullTarget string
}

func SolutionManager(solComms *SolutionComms, showPop bool) {
	var (
		block int
		diff  int
		sol   Solution
	)

	for {
		select {
		case newBlock := <-solComms.Block:
			// fmt.Println("Block is: ", block)
			if newBlock != block {
				block = newBlock
			}
		case _ = <-solComms.Step:
		case diff = <-solComms.Diff:
		case sol = <-solComms.Solution:
			// fmt.Printf("Solution is: %+v\n", sol)
			if sol.Block != block {
				// Drop stale solutions
				fmt.Printf("Dropping Solution (old block): %+v\n", sol)
				continue
			} else if sol.TargetLen <= sol.Chars-2 {
				// PoP solution
				if showPop {
					printFoundSolution(sol, false)
				}
			} else if sol.TargetLen == sol.Chars-1 && diff%10 == 0 {
				// PoP solution
				// When diff%10 == 0, there are no low steps
				if showPop {
					printFoundSolution(sol, false)
				}
			} else if sol.TargetLen == sol.Chars-1 && diff%10 != 0 {
				// Low step solution
				printFoundSolution(sol, true)
			} else {
				// High step solution
				printFoundSolution(sol, true)
			}
			solComms.SendChan <- fmt.Sprintf("STEP %d %s %s %d", sol.Block, sol.Seed, sol.HashStr, sol.TargetLen)
			solComms.StepSent <- struct{}{}
		}
	}
}

func printFoundSolution(sol Solution, isStep bool) {
	stepOrPop := "PoP"
	if isStep {
		stepOrPop = "Step"
	}
	fmt.Printf(
		found_one,
		stepOrPop,
		sol.Block,
		sol.Step,
		sol.Seed,
		sol.HashStr,
		sol.SolvedHash,
		sol.TargetLen,
		sol.Target,
		sol.FullTarget,
	)
}

const found_one = `************************************
FOUND %s SOLUTION
Block         : %d
Step          : %d
Seed          : %s
Hashed String : %s
SHA256 Value  : %s
Target Len    : %d
Target        : %s
Full Target   : %s
************************************
`
