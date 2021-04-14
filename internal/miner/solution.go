package miner

import "fmt"

func NewSolutionComms(sendChan chan string) *SolutionComms {
	return &SolutionComms{
		Block:    make(chan int, 10),
		Step:     make(chan int, 10),
		Diff:     make(chan int, 10),
		Solution: make(chan Solution, 10),
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
	Seed    string
	HashNum int
	Block   int
	Chars   int
	Step    int
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
			// fmt.Println("charsSml is: ", charsSml)
			// fmt.Println("charsLrg is: ", charsLrg)
			// fmt.Println("charsChg is: ", charsChg)
			// fmt.Println("charsCurr is: ", charsCurr)

			// Send any stored solutions that meet this criteria
			if sols, ok := storedSols[charsCurr]; ok {
				for _, s := range sols {
					solComms.SendChan <- fmt.Sprintf("STEP %d %s %d", s.Block, s.Seed, s.HashNum)
				}
				storedSols[charsCurr] = make([]Solution, 10)
			}
		case sol = <-solComms.Solution:
			fmt.Printf("Solution is: %+v\n", sol)
			if sol.Block != block {
				// Drop stale solutions
				fmt.Printf("Dropping Solution (old block): %+v", sol)
				continue
			} else if charsLrg-sol.Chars > 1 {
				// Drop solutions where the step has already dropped
				// and so these will never get sent
				fmt.Printf("Dropping Solution (low diff): %+v", sol)
				continue
			} else if step < charsChg && sol.Chars < charsLrg {
				// Found a future solution, store it
				storedSols[sol.Chars] = append(storedSols[sol.Chars], sol)
			} else {
				// Any solution that gets here is valid: send it
				solComms.SendChan <- fmt.Sprintf("STEP %d %s %d", sol.Block, sol.Seed, sol.HashNum)
			}
		}
	}
}
