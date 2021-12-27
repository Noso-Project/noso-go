package common

import (
	"errors"
	"fmt"
	"strconv"
)

type ServerMessageType int

const (
	JOIN ServerMessageType = iota + 1
	JOINOK
	PING
	PONG
	PASSFAILED
)

type ServerMessage interface {
	GetMsgType() ServerMessageType
}

type serverMessage struct {
	MsgType ServerMessageType
}

func (s serverMessage) GetMsgType() ServerMessageType {
	return s.MsgType
}

type joinOk struct {
	serverMessage
	poolAddr          string
	minerSeed         string
	block             int
	targetHash        string
	targetLen         int
	currentStep       int
	difficulty        int
	poolBalance       string
	blocksTillPayment int
	poolHashrate      int
	poolDepth         int
}

func newJoinOk(msg []string) (message joinOk) {
	// TODO: Handle strconv errors
	message = joinOk{}
	message.MsgType = JOINOK
	message.poolAddr = msg[1]
	message.minerSeed = msg[2]
	block, _ := strconv.Atoi(msg[4])
	message.block = block
	message.targetHash = msg[5]
	targetLen, _ := strconv.Atoi(msg[6])
	message.targetLen = targetLen
	step, _ := strconv.Atoi(msg[7])
	message.currentStep = step
	diff, _ := strconv.Atoi(msg[8])
	message.difficulty = diff
	message.poolBalance = msg[9]
	blocksTill, _ := strconv.Atoi(msg[10])
	message.blocksTillPayment = blocksTill
	poolHR, _ := strconv.Atoi(msg[11])
	message.poolHashrate = poolHR
	depth, _ := strconv.Atoi(msg[12])
	message.poolDepth = depth

	return message
}

type pong struct {
	serverMessage
	block             int
	targetHash        string
	targetLen         int
	currentStep       int
	difficulty        int
	poolBalance       string
	blocksTillPayment int
	poolHashrate      int
	poolDepth         int
}

func newPong(msg []string) (message pong) {
	// TODO: Handle strconv errors
	message = pong{}
	message.MsgType = PONG
	block, _ := strconv.Atoi(msg[2])
	message.block = block
	message.targetHash = msg[3]
	targetLen, _ := strconv.Atoi(msg[4])
	message.targetLen = targetLen
	step, _ := strconv.Atoi(msg[5])
	message.currentStep = step
	diff, _ := strconv.Atoi(msg[6])
	message.difficulty = diff
	message.poolBalance = msg[7]
	blocksTill, _ := strconv.Atoi(msg[8])
	message.blocksTillPayment = blocksTill
	poolHR, _ := strconv.Atoi(msg[9])
	message.poolHashrate = poolHR
	depth, _ := strconv.Atoi(msg[10])
	message.poolDepth = depth

	return message
}

type passFailed struct {
	serverMessage
}

func newPassFailed(msg []string) (message passFailed) {
	// TODO: Handle strconv errors
	message = passFailed{}
	message.MsgType = PASSFAILED

	return message
}

func stringToType(s string) (ServerMessageType, error) {
	switch s {
	case "JOIN":
		return JOIN, nil
	case "PING":
		return PING, nil
	default:
		return -1, errors.New(fmt.Sprint("Unknown req command: ", s))
	}
}
