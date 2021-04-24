package miner

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"
)

const (
	// Responses
	JOINOK_01     = "JOINOK JOINOK_01POOLADDRESS !3!!!!!!! PoolData 5891 JOINOK_01TARGETSTRING 11 0 102 0 -4 2215889"
	POOLSTEPS_01  = "POOLSTEPS PoolData 5890 POOLSTEPS_01TARGETSTRING 12 1 109 12345678 -5 2384383"
	PASSFAILED_01 = "PASSFAILED"
	PONG_01       = "PONG PoolData 5903 PONG_01TARGETSTRING 13 2 105 0 8 2364702"
	STEPOK_01     = "STEPOK"
)

func checkInt(actual, expected int) error {
	if actual != expected {
		return fmt.Errorf("Expected %d, found %d", expected, actual)
	}
	return nil
}

func checkString(actual, expected string) error {
	if actual != expected {
		return fmt.Errorf("Expected %s, found %s", expected, actual)
	}
	return nil
}

func NewParseTest() *ParseTest {
	resps := make(map[string]string)
	resps["JOINOK_01"] = JOINOK_01
	resps["POOLSTEPS_01"] = POOLSTEPS_01
	resps["PASSFAILED_01"] = PASSFAILED_01
	resps["PONG_01"] = PONG_01
	resps["STEPOK_01"] = STEPOK_01

	return &ParseTest{
		responses: resps,
	}
}

type ParseTest struct {
	responses map[string]string
	comms     *Comms
	poolIp    string
	wallet    string
	block     int
}

func (p *ParseTest) iHaveANewCommsObject() error {
	p.comms = NewComms()
	return nil
}

func (p *ParseTest) iHaveAPoolIPOf(poolIp string) error {
	p.poolIp = poolIp
	return nil
}

func (p *ParseTest) iHaveAWalletAddressOf(wallet string) error {
	p.wallet = wallet
	return nil
}

func (p *ParseTest) iParseTheResponse(resp string) error {
	if r, ok := p.responses[resp]; !ok {
		return fmt.Errorf("Could not match %s to a known response", resp)
	} else {
		go Parse(p.comms, p.poolIp, p.wallet, p.block, r)
		return nil
	}
}

func (p *ParseTest) iTheCurrentBlockIs(block int) error {
	p.block = block
	return nil
}

func (p *ParseTest) theCommsJoinedChannelShouldBeCalled() error {
	select {
	case <-p.comms.Joined:
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("Timed out after 3 seconds waiting for comms.Joined to be called")
	}
}

func (p *ParseTest) theCommsMinerSeedChannelShouldHave(expected string) error {
	return checkString(<-p.comms.MinerSeed, expected)
}

func (p *ParseTest) theCommsPoolAddrChannelShouldHave(expected string) error {
	return checkString(<-p.comms.PoolAddr, expected)
}
func (p *ParseTest) theCommsBalanceChannelShouldHave(expected string) error {
	return checkString(<-p.comms.Balance, expected)
}

func (p *ParseTest) theCommsBlockChannelShouldHave(expected int) error {
	return checkInt(<-p.comms.Block, expected)
}

func (p *ParseTest) theCommsBlocksTillPaymentChannelShouldHave(expectedStr string) error {
	expected, err := strconv.Atoi(expectedStr)
	if err != nil {
		return fmt.Errorf("Could not convert expected BTP value to int: %s", expectedStr)
	}

	return checkInt(<-p.comms.BlocksTillPayment, expected)
}

func (p *ParseTest) theCommsDiffChannelShouldHave(expected int) error {
	return checkInt(<-p.comms.Diff, expected)
}

func (p *ParseTest) theCommsStepChannelShouldHave(expected int) error {
	return checkInt(<-p.comms.Step, expected)
}

func (p *ParseTest) theCommsTargetCharsChannelShouldHave(expected int) error {
	return checkInt(<-p.comms.TargetChars, expected)
}

func (p *ParseTest) theCommsTargetStringChannelShouldHave(expected string) error {
	return checkString(<-p.comms.TargetString, expected)
}

func (p *ParseTest) theCommsStepSolvedChannelShouldHave(expected int) error {
	return checkInt(<-p.comms.StepSolved, expected)
}

func (p *ParseTest) noCommsChannelsGetCalled() error {
	chanName := ""
	select {
	case <-p.comms.PoolAddr:
		chanName = "comms.PoolAddr"
	case <-p.comms.MinerSeed:
		chanName = "comms.MinerSeed"
	case <-p.comms.TargetString:
		chanName = "comms.TargetString"
	case <-p.comms.TargetChars:
		chanName = "comms.TargetChars"
	case <-p.comms.Block:
		chanName = "comms.Block"
	case <-p.comms.Step:
		chanName = "comms.Step"
	case <-p.comms.Diff:
		chanName = "comms.Diff"
	case <-p.comms.Balance:
		chanName = "comms.Balance"
	case <-p.comms.BlocksTillPayment:
		chanName = "comms.BlocksTillPayment"
	case <-p.comms.StepSolved:
		chanName = "comms.StepSolved"
	case <-p.comms.HashRate:
		chanName = "comms.HashRate"
	case <-p.comms.Jobs:
		chanName = "comms.Jobs"
	case <-p.comms.Reports:
		chanName = "comms.Reports"
	case <-p.comms.Solutions:
		chanName = "comms.Solution"
	case <-p.comms.Joined:
		chanName = "comms.Joined"
	case <-time.After(250 * time.Millisecond):
	}

	if chanName != "" {
		return fmt.Errorf("Unexpectedly received on channel: %s", chanName)
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {

	p := NewParseTest()

	ctx.Step(`^I have a new comms object$`, p.iHaveANewCommsObject)
	ctx.Step(`^I have a pool IP of "([^"]*)"$`, p.iHaveAPoolIPOf)
	ctx.Step(`^I have a wallet address of "([^"]*)"$`, p.iHaveAWalletAddressOf)
	ctx.Step(`^I parse the "([^"]*)" response$`, p.iParseTheResponse)
	ctx.Step(`^I the current block is (\d+)$`, p.iTheCurrentBlockIs)
	ctx.Step(`^the comms\.Joined channel should be called$`, p.theCommsJoinedChannelShouldBeCalled)
	ctx.Step(`^the comms\.MinerSeed channel should have "([^"]*)"$`, p.theCommsMinerSeedChannelShouldHave)
	ctx.Step(`^the comms\.PoolAddr channel should have "([^"]*)"$`, p.theCommsPoolAddrChannelShouldHave)
	ctx.Step(`^the comms\.Balance channel should have "([^"]*)"$`, p.theCommsBalanceChannelShouldHave)
	ctx.Step(`^the comms\.Block channel should have (\d+)$`, p.theCommsBlockChannelShouldHave)
	ctx.Step(`^the comms\.BlocksTillPayment channel should have "([^"]*)"$`, p.theCommsBlocksTillPaymentChannelShouldHave)
	ctx.Step(`^the comms\.Diff channel should have (\d+)$`, p.theCommsDiffChannelShouldHave)
	ctx.Step(`^the comms\.Step channel should have (\d+)$`, p.theCommsStepChannelShouldHave)
	ctx.Step(`^the comms\.TargetChars channel should have (\d+)$`, p.theCommsTargetCharsChannelShouldHave)
	ctx.Step(`^the comms\.TargetString channel should have "([^"]*)"$`, p.theCommsTargetStringChannelShouldHave)
	ctx.Step(`^the comms\.StepSolved channel should have (\d+)$`, p.theCommsStepSolvedChannelShouldHave)
	ctx.Step(`^no comms channels get called$`, p.noCommsChannelsGetCalled)
}
