package common

import (
	"math/rand"
	"time"

	"github.com/denisbrodbeck/machineid"
)

var deviceId = getMachineId()
var instanceId = getInstanceId()

func getMachineId() string {
	id, err := machineid.ProtectedID("noso-go")
	if err != nil {
		// log.Error(err)
		id = "123456"
	}

	return id[:6]
}

// From https://stackoverflow.com/a/22892986/4079962
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func getInstanceId() string {
	b := make([]rune, 6)
	for i := range b {
		rand.Seed(time.Now().UnixNano())
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
