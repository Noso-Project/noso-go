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

func NewComms() *Comms {
	return &Comms{
		PoolAddr:     make(chan string, 0),
		MinerSeed:    make(chan string, 0),
		TargetBlock:  make(chan int, 0),
		TargetString: make(chan string, 0),
		TargetChars:  make(chan int, 0),
		CurrentStep:  make(chan int, 0),
		Hashes:       make(chan int, 0),
	}
}
