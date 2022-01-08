package common

import (
	"strconv"
	"time"
)

func NewHashRateReport(name string, hashes int, start, stop time.Time) HashRateReport {
	duration := stop.Sub(start)
	perSec := float64(duration) / float64(time.Second)
	hr := int(float64(hashes) / perSec)
	return HashRateReport{
		MinerName: name,
		HashRate:  hr,
		hashes:    hashes,
		start:     start,
		stop:      stop,
		duration:  duration,
	}
}

type HashRateReport struct {
	MinerName   string
	HashRate    int
	hashes      int
	start, stop time.Time
	duration    time.Duration
}

func (h HashRateReport) String() string {
	return FormatHashRate(strconv.Itoa(h.HashRate))
}
