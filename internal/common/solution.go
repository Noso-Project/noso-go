package common

import "fmt"

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

func (s Solution) String() string {
	return fmt.Sprintf("STEP %d %s %s %d %s", s.Block, s.Seed, s.HashStr, s.TargetLen, instanceId)
}
