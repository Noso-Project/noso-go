package common

import "strconv"

type ServerMessageType int

const (
	JOINOK ServerMessageType = iota + 1
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
