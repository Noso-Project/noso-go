package common

import (
	"errors"
	"strings"
)

var (
	EmptyRespErr   = errors.New("Server sent an empty response")
	UnknownRespErr = errors.New("Server sent an unknown response type")
)

func parse(msg string) (ServerMessage, error) {
	if len(msg) == 0 {
		return serverMessage{}, EmptyRespErr
	}

	splitMsg := strings.Split(msg, " ")
	switch splitMsg[0] {
	case "JOINOK":
		return newJoinOk(splitMsg), nil
	case "PONG":
		return newPong(splitMsg), nil
	case "PASSFAILED":
		return newPassFailed(splitMsg), nil
	case "POOLSTEPS":
		return newPoolSteps(splitMsg), nil
	default:
		return serverMessage{}, UnknownRespErr
	}
}
