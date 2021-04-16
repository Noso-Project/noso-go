package miner

import "fmt"

func NewSolutionComms(sendChan chan string) *SolutionComms {
	return &SolutionComms{
		Block:    make(chan int, 0),
		Step:     make(chan int, 0),
		Diff:     make(chan int, 0),
		Solution: make(chan Solution, 0),
		SendChan: sendChan,
	}
}

type SolutionComms struct {
	Block    chan int
	Step     chan int
	Diff     chan int
	Solution chan Solution
	SendChan chan string
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

func SolutionManager(solComms *SolutionComms) {
	var (
		block      int
		step       int
		diff       int
		charsCurr  int // Current chars, based on step and diff
		charsChg   int // The step where char len change from lrg to sml
		charsLrg   int // The largest possible char size for the diff
		charsSml   int // The smallest possible char size for the diff
		sol        Solution
		storedSols map[int][]Solution
	)

	for {
		select {
		case newBlock := <-solComms.Block:
			// fmt.Println("Block is: ", block)
			if newBlock != block {
				storedSols = make(map[int][]Solution)
				block = newBlock
			}
		case step = <-solComms.Step:
			// fmt.Println("Step is: ", step)
		case diff = <-solComms.Diff:
			// fmt.Println("Diff is: ", diff)
			// At diff 90, all steps have target chars 9
			// At diff 95, steps 0 - 4 have target chars 10, 5 - 9 have 9
			charsSml = diff / 10
			charsLrg = charsSml + 1
			charsChg = diff % 10
			if step < charsChg {
				charsCurr = charsLrg
			} else {
				charsCurr = charsSml
			}

			// Send any stored solutions that meet this criteria
			if sols, ok := storedSols[charsCurr]; ok {
				drop := 0
				for idx, s := range sols {
					if idx < 10 {
						solComms.SendChan <- fmt.Sprintf("STEP %d %s %s", s.Block, s.Seed, s.HashStr)
					} else {
						drop++
					}
				}
				if drop > 0 {
					fmt.Println("Already sent 10 stored responses, dropping the remaining %d stored solutions")
				}
				storedSols[charsCurr] = make([]Solution, 0)
			}
		case sol = <-solComms.Solution:
			fmt.Printf("Solution is: %+v\n", sol)
			if sol.Block != block {
				// Drop stale solutions
				fmt.Printf("Dropping Solution (old block): %+v\n", sol)
				continue
			} else if charsLrg-sol.Chars > 1 {
				// Drop solutions where the step difficulty has already dropped
				// and so these will never get sent. For instance, if diff is 95
				// and we are on step 6, the miner will find solutions that are 8
				// chars long, but those will never be valid
				fmt.Printf("Dropping Solution (low diff): %+v\n", sol)
				continue
			} else if step < charsChg && sol.Chars < charsLrg {
				// Found a future solution, store it
				// Should only store 10 max
				// TODO: Max number of future solutions that can be accepted
				//       varies with the difficulty. Should only store max
				//       possible and not just a blanket 10
				if len(storedSols[sol.Chars]) < 10 {
					storedSols[sol.Chars] = append(storedSols[sol.Chars], sol)
					fmt.Println("Found a future solution, storing")
					printFoundSolution(sol, true)
					fmt.Printf("Currently have %d solutions stored\n", len(storedSols[sol.Chars]))
				} else {
					fmt.Printf("10 future solutions already stored, dropping: %+v\n", sol)
				}
			} else {
				// Any solution that gets here is valid: send it
				solComms.SendChan <- fmt.Sprintf("STEP %d %s %s", sol.Block, sol.Seed, sol.HashStr)
				if sol.Target == sol.FullTarget {
					printFoundSolution(sol, false)
				}
			}
		}
	}
}

func printFoundSolution(sol Solution, future bool) {
	currentOrFuture := "CURRENT"
	storeOrSend := "Sending"
	if future {
		currentOrFuture = "FUTURE"
		storeOrSend = "Storing"
	}
	fmt.Printf(
		found_one,
		currentOrFuture,
		storeOrSend,
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
Store or Send : %s
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
