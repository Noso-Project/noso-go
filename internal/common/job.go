package common

import "context"

func NewJob(ctx context.Context) Job {
	return Job{
		done: ctx.Done(),
	}
}

type Job struct {
	done          <-chan struct{}
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

func (j *Job) Done() <-chan struct{} {
	return j.done
}

type JobStreamReq struct {
	Stream chan (<-chan Job)
}
