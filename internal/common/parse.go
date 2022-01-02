package common

import (
	"errors"
	"strings"
)

var (
	ErrEmptyResp   = errors.New("Server sent an empty response")
	ErrUnknownResp = errors.New("Server sent an unknown response type")
)

func parse(msg string) (ServerMessage, error) {
	if len(msg) == 0 {
		return serverMessage{}, ErrEmptyResp
	}

	splitMsg := strings.Split(msg, " ")
	switch splitMsg[0] {
	case "JOINOK":
		return newJoinOk(splitMsg), nil
	case "PONG":
		return newPong(splitMsg), nil
	case "PASSFAILED":
		return newPassFailed(splitMsg), nil
	case "ALREADYCONNECTED":
		return newAlreadyConnected(splitMsg), nil
	case "POOLSTEPS":
		return newPoolSteps(splitMsg), nil
	case "STEPOK":
		return newStepOk(splitMsg), nil
	default:
		return serverMessage{}, ErrUnknownResp
	}
}
