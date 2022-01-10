package miner

import (
	"context"
	"sync"

	"github.com/Noso-Project/noso-go/internal/common"
)

func Run(opts common.Opts) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker := common.NewBroker(ctx)
	client := common.NewClient(ctx, broker, opts)

	var wg sync.WaitGroup

	wg.Add(3)
	go JobManager(ctx, client, broker, &wg)
	go SolutionManager(ctx, client, broker, &wg)
	go MinerManagerNew(ctx, client, broker, opts, &wg)
	wg.Wait()

	err := client.Connect()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
	}

	return nil
}
