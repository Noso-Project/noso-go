package common

import (
	"errors"
	"fmt"
	"strconv"
)

type ServerMessageType int

func (s ServerMessageType) String() string {
	// TODO: Seems like I should be able to do this without a switch statement
	switch s {
	case JOIN:
		return "JOIN"
	case JOINOK:
		return "JOINOK"
	case PING:
		return "PING"
	case PONG:
		return "PONG"
	case POOLSTEPS:
		return "POOLSTEPS"
	case PASSFAILED:
		return "PASSFAILED"
	case ALREADYCONNECTED:
		return "ALREADYCONNECTED"
	case STEP:
		return "STEP"
	case STEPOK:
		return "STEPOK"
	default:
		return fmt.Sprintf("%d (cant find string)", int(s))
	}
}

const (
	JOIN ServerMessageType = iota + 1
	JOINOK
	PING
	PONG
	POOLSTEPS
	PASSFAILED
	ALREADYCONNECTED
	STEP
	STEPOK
	OTHER
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

type JoinOk struct {
	serverMessage
	PoolAddr          string
	MinerSeed         string
	Block             int
	TargetHash        string
	TargetLen         int
	CurrentStep       int
	Difficulty        int
	PoolBalance       string
	BlocksTillPayment int
	PoolHashrate      int
	PoolDepth         int
}

func newJoinOk(msg []string) (message JoinOk) {
	// TODO: Handle strconv errors
	// TODO: Check msg has expected len before indexing
	message = JoinOk{}
	message.MsgType = JOINOK
	message.PoolAddr = msg[1]
	message.MinerSeed = msg[2]
	block, _ := strconv.Atoi(msg[4])
	message.Block = block
	message.TargetHash = msg[5]
	targetLen, _ := strconv.Atoi(msg[6])
	message.TargetLen = targetLen
	step, _ := strconv.Atoi(msg[7])
	message.CurrentStep = step
	diff, _ := strconv.Atoi(msg[8])
	message.Difficulty = diff
	message.PoolBalance = msg[9]
	blocksTill, _ := strconv.Atoi(msg[10])
	message.BlocksTillPayment = blocksTill
	poolHR, _ := strconv.Atoi(msg[11])
	message.PoolHashrate = poolHR
	depth, _ := strconv.Atoi(msg[12])
	message.PoolDepth = depth

	return message
}

type Pong struct {
	serverMessage
	Block             int
	TargetHash        string
	TargetLen         int
	CurrentStep       int
	Difficulty        int
	PoolBalance       string
	BlocksTillPayment int
	PoolHashrate      int
	PoolDepth         int
}

func newPong(msg []string) (message Pong) {
	// TODO: Handle strconv errors
	// TODO: Check msg has expected len before indexing
	message = Pong{}
	message.MsgType = PONG
	block, _ := strconv.Atoi(msg[2])
	message.Block = block
	message.TargetHash = msg[3]
	targetLen, _ := strconv.Atoi(msg[4])
	message.TargetLen = targetLen
	step, _ := strconv.Atoi(msg[5])
	message.CurrentStep = step
	diff, _ := strconv.Atoi(msg[6])
	message.Difficulty = diff
	message.PoolBalance = msg[7]
	blocksTill, _ := strconv.Atoi(msg[8])
	message.BlocksTillPayment = blocksTill
	poolHR, _ := strconv.Atoi(msg[9])
	message.PoolHashrate = poolHR
	depth, _ := strconv.Atoi(msg[10])
	message.PoolDepth = depth

	return message
}

type PoolSteps struct {
	serverMessage
	Block             int
	TargetHash        string
	TargetLen         int
	CurrentStep       int
	Difficulty        int
	PoolBalance       string
	BlocksTillPayment int
	PoolHashrate      int
	PoolDepth         int
}

func newPoolSteps(msg []string) (message PoolSteps) {
	// TODO: Handle strconv errors
	// TODO: Check msg has expected len before indexing
	message = PoolSteps{}
	message.MsgType = POOLSTEPS
	block, _ := strconv.Atoi(msg[2])
	message.Block = block
	message.TargetHash = msg[3]
	targetLen, _ := strconv.Atoi(msg[4])
	message.TargetLen = targetLen
	step, _ := strconv.Atoi(msg[5])
	message.CurrentStep = step
	diff, _ := strconv.Atoi(msg[6])
	message.Difficulty = diff
	message.PoolBalance = msg[7]
	blocksTill, _ := strconv.Atoi(msg[8])
	message.BlocksTillPayment = blocksTill
	poolHR, _ := strconv.Atoi(msg[9])
	message.PoolHashrate = poolHR
	depth, _ := strconv.Atoi(msg[10])
	message.PoolDepth = depth

	return message
}

type PassFailed struct {
	serverMessage
}

func newPassFailed(msg []string) (message PassFailed) {
	// TODO: Handle strconv errors
	// TODO: Check msg has expected len before indexing
	message = PassFailed{}
	message.MsgType = PASSFAILED

	return message
}

type AlreadyConnected struct {
	serverMessage
}

func newAlreadyConnected(msg []string) (message AlreadyConnected) {
	// TODO: Handle strconv errors
	// TODO: Check msg has expected len before indexing
	message = AlreadyConnected{}
	message.MsgType = ALREADYCONNECTED

	return message
}

type StepOk struct {
	serverMessage
	PopValue int
}

func newStepOk(msg []string) (message StepOk) {
	// TODO: Handle strconv errors
	// TODO: Check msg has expected len before indexing
	message = StepOk{}
	message.MsgType = STEPOK
	popValue, _ := strconv.Atoi(msg[1])
	message.PopValue = popValue

	return message
}

func stringToType(s string) (ServerMessageType, error) {
	switch s {
	case "JOIN":
		return JOIN, nil
	case "PING":
		return PING, nil
	case "STEP":
		return STEP, nil
	default:
		return -1, errors.New(fmt.Sprint("Unknown req command: ", s))
	}
}
