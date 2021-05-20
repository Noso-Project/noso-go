package miner

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const CSVHEADER = "Transaction Time,Pool IP Address,Wallet Address,Request Or Response,Block,Payment Amount,Order Id\n"

func CreateLogPaymentsFile() {
	write("")
}

func LogPaymentReq(poolIp string, wallet string, block int, amount string) {
	var (
		now      time.Time
		writeStr string
	)
	now = time.Now()

	amount = parseAmount(amount)

	writeStr = fmt.Sprintf("%s,%s,%s,%s,%d,%s,\n", now.Format(time.RFC3339), poolIp, wallet, "Payment Request", block, amount)

	write(writeStr)
}

func LogPaymentResp(paymentMsg []string, poolIp string) {
	var (
		now      time.Time
		amount   string
		orderId  string
		writeStr string
		wallet   string
	)

	if len(paymentMsg) < 3 {
		fmt.Println("Error: noso-go requires that the pool use Noso Wallet 0.2.0 N or greater")
		return
	}

	// Example PAYMENTOK response
	// <- PAYMENTOK 1618891646 POOLIP Nm6jiGfRg7DVHHMfbMJL9CT1DtkUCF 2 5833 1.48153045 OR60v3w4j25pkl7mp2aaxa6l7g7hxqsdlfu86fkueh11tfyqg03z
	// <- PAYMENTOK [TIMESTAMP] POOLIP [MINERADDRESS] 2 [BLOCK] [AMOUNT] [ORDERID]
	// Format payment into X.YYY format

	wallet = paymentMsg[3]
	block, err := strconv.Atoi(paymentMsg[5])
	if err != nil {
		fmt.Println("Error converting block string in PAYMENTOK response: ", err)
	}
	amount = paymentMsg[6]
	orderId = paymentMsg[7]

	now = time.Now()

	amount = parseAmount(amount)

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
			if _, err := f.WriteString(CSVHEADER); err != nil {
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

func parseAmount(amount string) string {
	var (
		whole string
		frac  string
	)
	l := len(amount)
	if amount == "0" {
		amount = "0.00000000"
	} else if strings.Contains(amount, ".") {
		// Already formatted to X.YYY, do nothing
	} else if l == 8 {
		amount = "0." + amount
	} else if l > 8 {
		whole = amount[:l-8]
		frac = amount[l-8:]
		amount = fmt.Sprintf("%s.%s", whole, frac)
	} else {
		amount = fmt.Sprintf("0.%08s", l)
	}
	return amount
}
