package common

type Job struct {
	PoolAddr      string
	MinerSeedBase string
	MinerSeed     string
	TargetString  string
	TargetChars   int
	Diff          int
	Block         int
	Step          int
	PoolDepth     int
}

type JobStreamReq struct {
	Stream chan (<-chan Job)
}
