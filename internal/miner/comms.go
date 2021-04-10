package miner

type Comms struct {
	CurrentBlock chan int
	TargetBlock  chan int
	CurrentStep  chan int
	Hashes       chan int
}
