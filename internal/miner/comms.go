package miner

type Comms struct {
	PoolAddr     chan string
	MinerSeed    chan string
	TargetBlock  chan int
	TargetString chan string
	TargetChars  chan int
	CurrentStep  chan int
	Hashes       chan int
}
