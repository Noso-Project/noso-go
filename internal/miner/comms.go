package miner

func NewComms() *Comms {
	return &Comms{
		PoolAddr:     make(chan string, 10),
		MinerSeed:    make(chan string, 10),
		TargetString: make(chan string, 10),
		TargetChars:  make(chan int, 10),
		Block:        make(chan int, 10),
		Step:         make(chan int, 10),
		Diff:         make(chan int, 10),
		StepSolved:   make(chan int, 10),
		HashRate:     make(chan int, 10),
		Jobs:         make(chan Job, 10),
		Reports:      make(chan Report, 10),
		Solutions:    make(chan Solution, 10),
		Joined:       make(chan struct{}, 10),
	}
}

type Comms struct {
	PoolAddr     chan string
	MinerSeed    chan string
	TargetString chan string
	TargetChars  chan int
	Block        chan int
	Step         chan int
	Diff         chan int
	StepSolved   chan int
	HashRate     chan int
	Jobs         chan Job
	Reports      chan Report
	Solutions    chan Solution
	Joined       chan struct{}
}

type Report struct {
	WorkerNum string
	Hashes    int
}
