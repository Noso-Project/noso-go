package miner

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Noso-Project/noso-go/internal/common"
)

func ReportManager(ctx context.Context, broker *common.Broker, wg *sync.WaitGroup) {
	wg.Done()

	hashReportStream, err := broker.Subscribe(ctx, common.HashRateTopic)
	if err != nil {
		panic(fmt.Sprint("Got an err and didn't expect one", err))
	}

	reports := make(map[string]common.HashRateReport)
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case h, ok := <-hashReportStream:
			if !ok {
				// TODO: Need to think about this more
				panic("hashReportStream died unexpectidly")
				return
			}
			switch h.(type) {
			case common.HashRateReport:
				reports[h.(common.HashRateReport).MinerName] = h.(common.HashRateReport)
			}
		case <-ticker.C:
			hr := 0
			minerCount := 0
			for _, v := range reports {
				minerCount++
				hr += v.HashRate
			}
			if minerCount == 0 {
				break
			}
			totalHr := common.FormatHashRate(strconv.Itoa(hr))
			minerAve := common.FormatHashRate(strconv.Itoa(int(hr / minerCount)))
			fmt.Printf("HashRate -> %s Total, %s Average (%d miners)\n", totalHr, minerAve, minerCount)
			broker.Publish(ctx, common.NewTotalHashRateReport(hr))
		}
	}
}
