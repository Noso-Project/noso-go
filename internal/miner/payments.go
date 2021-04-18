package miner

import (
	"fmt"
	"os"
	"time"
)

func logPayment(paymentMsg []string) {
	var (
		now      time.Time
		amount   string
		orderId  string
		whole    string
		frac     string
		writeStr string
	)
	now = time.Now()
	amount = paymentMsg[1]
	if len(paymentMsg) > 2 {
		orderId = paymentMsg[2]
	}

	// Format payment into X.YYY format
	l := len(amount)
	if amount == "0" {
		amount = "0.00000000"
	} else if l == 8 {
		amount = "0." + amount
	} else {
		whole = amount[:l-8]
		frac = amount[l-8:]
		amount = fmt.Sprintf("%s.%s", whole, frac)
	}

	writeStr = fmt.Sprintf("%s,%s,%s\n", now, amount, orderId)

	f, err := os.OpenFile("payments.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Printf("Trouble opening payments.csv: %s\n", err)
		return
	}

	defer f.Close()

	if _, err2 := f.WriteString(writeStr); err2 != nil {
		fmt.Printf("Trouble writing to payments.csv: %s\n", err2)
	}
}
