package miner

import (
	"context"
	"sync"

	"github.com/Noso-Project/noso-go/internal/common"
)

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	broker := common.NewBroker(ctx)
	// client := common.NewClient(ctx, broker, "pool.noso.dev", 8082)
	client := common.NewClient(ctx, broker, "devnosoeu.nosocoin.com", 8082)

	var wg sync.WaitGroup

	wg.Add(3)
	go JobManager(ctx, client, broker, &wg)
	go SolutionManager(ctx, client, broker, &wg)
	go MinerManager(ctx, client, broker, &wg)
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

func RunMiner(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	broker := common.NewBroker(ctx)
	// client := common.NewClient(ctx, broker, "pool.noso.dev", 8082)
	client := common.NewClient(ctx, broker, "164.90.252.232", 8082)
	// client := common.NewClient(ctx, broker, "devnosoeu.nosocoin.com", 8082)

	var wg sync.WaitGroup

	wg.Add(3)
	go JobManager(ctx, client, broker, &wg)
	go SolutionManager(ctx, client, broker, &wg)
	go MinerManagerNew(ctx, client, broker, &wg)
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
