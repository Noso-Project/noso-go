package miner

import (
	"fmt"
	"math"
)

var (
	hashOrders = []string{"H", "H", "H", "H", "Kh", "Kh", "Kh", "Mh", "Mh", "Mh", "Gh", "Gh", "Gh", "Th", "Th", "Th", "Ph", "Ph", "Ph", "Eh", "Eh", "Eh", "Zh", "Zh", "Zh"}
)

func formatHashRate(hr string) string {
	mag := len(hr)

	if len(hr) < 4 {
		return fmt.Sprintf("%3s     %sash/s", hr, hashOrders[mag])
	}

	splitAt := math.Mod(float64(mag), 3)

	if splitAt == 0 {
		splitAt = 3
	}

	whole, frac := hr[:int(splitAt)], hr[int(splitAt):int(splitAt)+3]

	return fmt.Sprintf("%3s.%s %sash/s", whole, frac, hashOrders[mag])
}
