package common

import (
	"testing"
)

const (
	PASSFAILED_default       = "PASSFAILED"
	ALREADYCONNECTED_default = "ALREADYCONNECTED"
	JOINOK_default           = "JOINOK N6VxgLSpbni8kLbyUAjYXdHCPt2VEp 020000000 PoolData 37873 E1151A4F79E6394F6897A913ADCD476B 11 0 102 0 -30 42270 3"
	PONG_default             = "PONG PoolData 37892 C74B9ABA60E2EE1B52613959D4F06876 11 0 105 0 -29 86070 3"
	POOLSTEPS_default        = "POOLSTEPS PoolData 38441 AD23A982B87D193E8384EB50C3F0B50C 11 0 106 0 -23 43328 3"
	STEPOK_default           = "STEPOK 256"
)

func TestParse(t *testing.T) {
	t.Run("empty resp", func(t *testing.T) {
		_, got := parse("")
		want := ErrEmptyResp

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("unknown resp", func(t *testing.T) {
		_, got := parse("FOO BAR BAZ")
		want := ErrUnknownResp

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("joinOk", func(t *testing.T) {
		resp, err := parse(JOINOK_default)

		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err.Error())
		}

		got := resp.GetMsgType()
		want := JOINOK

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		assertMsgAttrs(t, resp.(JoinOk).PoolAddr, "N6VxgLSpbni8kLbyUAjYXdHCPt2VEp")
		assertMsgAttrs(t, resp.(JoinOk).MinerSeed, "020000000")
		assertMsgAttrs(t, resp.(JoinOk).Block, 37873)
		assertMsgAttrs(t, resp.(JoinOk).TargetHash, "E1151A4F79E6394F6897A913ADCD476B")
		assertMsgAttrs(t, resp.(JoinOk).TargetLen, 11)
		assertMsgAttrs(t, resp.(JoinOk).CurrentStep, 0)
		assertMsgAttrs(t, resp.(JoinOk).Difficulty, 102)
		assertMsgAttrs(t, resp.(JoinOk).PoolBalance, "0")
		assertMsgAttrs(t, resp.(JoinOk).BlocksTillPayment, -30)
		assertMsgAttrs(t, resp.(JoinOk).PoolHashrate, 42270)
		assertMsgAttrs(t, resp.(JoinOk).PoolDepth, 3)
	})
	t.Run("pong", func(t *testing.T) {
		resp, err := parse(PONG_default)

		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err.Error())
		}

		got := resp.GetMsgType()
		want := PONG

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		assertMsgAttrs(t, resp.(Pong).Block, 37892)
		assertMsgAttrs(t, resp.(Pong).TargetHash, "C74B9ABA60E2EE1B52613959D4F06876")
		assertMsgAttrs(t, resp.(Pong).TargetLen, 11)
		assertMsgAttrs(t, resp.(Pong).CurrentStep, 0)
		assertMsgAttrs(t, resp.(Pong).Difficulty, 105)
		assertMsgAttrs(t, resp.(Pong).PoolBalance, "0")
		assertMsgAttrs(t, resp.(Pong).BlocksTillPayment, -29)
		assertMsgAttrs(t, resp.(Pong).PoolHashrate, 86070)
		assertMsgAttrs(t, resp.(Pong).PoolDepth, 3)
	})
	t.Run("poolSteps", func(t *testing.T) {
		resp, err := parse(POOLSTEPS_default)

		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err.Error())
		}

		got := resp.GetMsgType()
		want := POOLSTEPS

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		assertMsgAttrs(t, resp.(PoolSteps).Block, 38441)
		assertMsgAttrs(t, resp.(PoolSteps).TargetHash, "AD23A982B87D193E8384EB50C3F0B50C")
		assertMsgAttrs(t, resp.(PoolSteps).TargetLen, 11)
		assertMsgAttrs(t, resp.(PoolSteps).CurrentStep, 0)
		assertMsgAttrs(t, resp.(PoolSteps).Difficulty, 106)
		assertMsgAttrs(t, resp.(PoolSteps).PoolBalance, "0")
		assertMsgAttrs(t, resp.(PoolSteps).BlocksTillPayment, -23)
		assertMsgAttrs(t, resp.(PoolSteps).PoolHashrate, 43328)
		assertMsgAttrs(t, resp.(PoolSteps).PoolDepth, 3)
	})
	t.Run("stepOk", func(t *testing.T) {
		resp, err := parse(STEPOK_default)

		if err != nil {
			t.Fatal("Got an error and didn't expect one: ", err.Error())
		}

		got := resp.GetMsgType()
		want := STEPOK

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}

		assertMsgAttrs(t, resp.(StepOk).PopValue, 256)
	})
}

func assertMsgType(t *testing.T, got, want serverMessage) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}

}

func assertMsgAttrs(t *testing.T, got, want interface{}) {
	t.Helper()

	switch got.(type) {
	case string:
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	case int:
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	}

}
