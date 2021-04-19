package miner

import (
	"fmt"
	"os"
	"time"
)

const HEADER = "Transaction Time,Pool IP Address,Wallet Address,Request Or Response,Block,Payment Amount,Order Id\n"

func CreateLogPaymentsFile() {
	write("")
}

func LogPaymentReq(poolIp string, wallet string, block int) {
	var (
		now      time.Time
		writeStr string
	)
	now = time.Now()

	writeStr = fmt.Sprintf("%s,%s,%s,%s,%d,,\n", now.Format(time.RFC3339), poolIp, wallet, "Payment Request", block)

	write(writeStr)
}

func LogPaymentResp(paymentMsg []string, poolIp, wallet string, block int) {
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

	writeStr = fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s\n", now.Format(time.RFC3339), poolIp, wallet, "Payment Response", block, amount, orderId)

	write(writeStr)
}

func write(writeStr string) {
	f, err := os.OpenFile("payments.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Printf("Trouble opening payments.csv: %s\n", err)
		return
	}

	defer f.Close()

	// Check to see if the file size is 0, indicating we just created it.
	// If we did create it, add csv headers first
	s, err := f.Stat()

	if err != nil {
		fmt.Printf("Trouble getting file stats for payments.csv: %s\n", err)
	} else {
		size := s.Size()
		if size == 0 {
			if _, err := f.WriteString(HEADER); err != nil {
				fmt.Printf("Trouble header to payments.csv: %s\n", err)
			}
		}
	}

	// Special case: if writeString is empty it means we should only write the header
	if writeStr == "" {
		return
	}

	if _, err := f.WriteString(writeStr); err != nil {
		fmt.Printf("Trouble writing to payments.csv: %s\n", err)
	}
}
