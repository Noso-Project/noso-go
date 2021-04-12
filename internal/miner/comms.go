package miner

func NewComms() *Comms {
	return &Comms{
		PoolAddr:     make(chan string, 0),
		MinerSeed:    make(chan string, 0),
		TargetBlock:  make(chan int, 0),
		TargetString: make(chan string, 0),
		TargetChars:  make(chan int, 0),
		CurrentStep:  make(chan int, 0),
		StepSolved:   make(chan int, 0),
		HashRate:     make(chan int, 0),
		Jobs:         make(chan Job, 0),
		Reports:      make(chan Report, 0),
		Solutions:    make(chan Solution, 0),
		Pong:         make(chan interface{}, 0),
		NewChars:     make(chan int, 0),
		NewBlock:     make(chan int, 0),
		NewStep:      make(chan int, 0),
		NewString:    make(chan string, 0),
		NewPoolAddr:  make(chan string, 0),
		NewMinerSeed: make(chan string, 0),
	}
}

type Comms struct {
	PoolAddr     chan string
	MinerSeed    chan string
	TargetBlock  chan int
	TargetString chan string
	TargetChars  chan int
	CurrentStep  chan int
	StepSolved   chan int
	HashRate     chan int
	Jobs         chan Job
	Reports      chan Report
	Solutions    chan Solution
	Pong         chan interface{}
	Joined       chan interface{}

	// For communicating with the jobFeeder goroutine
	NewChars     chan int
	NewBlock     chan int
	NewStep      chan int
	NewString    chan string
	NewPoolAddr  chan string
	NewMinerSeed chan string
}

type Job struct {
	Start, Stop  int
	TargetChars  int
	TargetBlock  int
	TargetString string
	Seed         string
	PoolAddr     string
	Block        int
	Step         int
}

type Solution struct {
	Seed        string
	HashNum     int
	TargetBlock int
	TargetChars int
}

type Report struct {
	WorkerNum string
	Hashes    int
}
