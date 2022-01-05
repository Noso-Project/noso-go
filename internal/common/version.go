package common

import "fmt"

var (
	Commit    = "0000000"
	Version   = "v0.0.0"
	MinerName = fmt.Sprintf("ng%s", Version[1:])
)
