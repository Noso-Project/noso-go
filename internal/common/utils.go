package common

import (
	"fmt"
	"math"
	"strings"
)

var orderOfMag = []string{"H", "Kh", "Mh", "Gh", "Th", "Ph", "Eh", "Zh"}

func FormatHashRate(hr string) string {
	var (
		whole, frac string
	)
	if len(hr) == 0 {
		return ""
	}

	mag := orderOfMag[(len(hr)-1)/3]

	if len(hr) < 4 {
		whole, frac = hr, "000"
	} else {
		// Split the hashrate of 12345600000000000000000
		// into 12.345
		splitAt := int(math.Mod(float64(len(hr)), 3))

		if splitAt == 0 {
			splitAt = 3
		}

		whole, frac = hr[:splitAt], hr[splitAt:splitAt+3]
	}

	return fmt.Sprintf("%3s.%s %sash/s", whole, frac, mag)
}

func FormatBalance(balance string) string {
	return fmt.Sprintf("%s Noso", ParseAmount(balance))
}

func ParseAmount(amount string) string {
	if strings.Contains(amount, ".") {
		// Already formatted to X.YYY, do nothing
		return amount
	}

	for len(amount) < 9 {
		amount = "0" + amount
	}

	l := len(amount)
	return fmt.Sprintf("%s.%s", amount[:l-8], amount[l-8:])
}
