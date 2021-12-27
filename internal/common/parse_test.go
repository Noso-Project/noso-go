package common

import (
	"testing"
)

const (
	JOINOK_default     = "JOINOK N6VxgLSpbni8kLbyUAjYXdHCPt2VEp 020000000 PoolData 37873 E1151A4F79E6394F6897A913ADCD476B 11 0 102 0 -30 42270 3"
	PONG_default       = "PONG PoolData 37892 C74B9ABA60E2EE1B52613959D4F06876 11 0 105 0 -29 86070 3"
	PASSFAILED_default = "PASSFAILED"
	POOLSTEPS_default  = "POOLSTEPS PoolData 38441 AD23A982B87D193E8384EB50C3F0B50C 11 0 106 0 -23 43328 3"
)

func TestParse(t *testing.T) {
	t.Run("empty resp", func(t *testing.T) {
		_, got := parse("")
		want := EmptyRespErr

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("unknown resp", func(t *testing.T) {
		_, got := parse("FOO BAR BAZ")
		want := UnknownRespErr

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

		assertMsgAttrs(t, resp.(joinOk).poolAddr, "N6VxgLSpbni8kLbyUAjYXdHCPt2VEp")
		assertMsgAttrs(t, resp.(joinOk).minerSeed, "020000000")
		assertMsgAttrs(t, resp.(joinOk).block, 37873)
		assertMsgAttrs(t, resp.(joinOk).targetHash, "E1151A4F79E6394F6897A913ADCD476B")
		assertMsgAttrs(t, resp.(joinOk).targetLen, 11)
		assertMsgAttrs(t, resp.(joinOk).currentStep, 0)
		assertMsgAttrs(t, resp.(joinOk).difficulty, 102)
		assertMsgAttrs(t, resp.(joinOk).poolBalance, "0")
		assertMsgAttrs(t, resp.(joinOk).blocksTillPayment, -30)
		assertMsgAttrs(t, resp.(joinOk).poolHashrate, 42270)
		assertMsgAttrs(t, resp.(joinOk).poolDepth, 3)
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

		assertMsgAttrs(t, resp.(pong).block, 37892)
		assertMsgAttrs(t, resp.(pong).targetHash, "C74B9ABA60E2EE1B52613959D4F06876")
		assertMsgAttrs(t, resp.(pong).targetLen, 11)
		assertMsgAttrs(t, resp.(pong).currentStep, 0)
		assertMsgAttrs(t, resp.(pong).difficulty, 105)
		assertMsgAttrs(t, resp.(pong).poolBalance, "0")
		assertMsgAttrs(t, resp.(pong).blocksTillPayment, -29)
		assertMsgAttrs(t, resp.(pong).poolHashrate, 86070)
		assertMsgAttrs(t, resp.(pong).poolDepth, 3)
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

		assertMsgAttrs(t, resp.(poolSteps).block, 38441)
		assertMsgAttrs(t, resp.(poolSteps).targetHash, "AD23A982B87D193E8384EB50C3F0B50C")
		assertMsgAttrs(t, resp.(poolSteps).targetLen, 11)
		assertMsgAttrs(t, resp.(poolSteps).currentStep, 0)
		assertMsgAttrs(t, resp.(poolSteps).difficulty, 106)
		assertMsgAttrs(t, resp.(poolSteps).poolBalance, "0")
		assertMsgAttrs(t, resp.(poolSteps).blocksTillPayment, -23)
		assertMsgAttrs(t, resp.(poolSteps).poolHashrate, 43328)
		assertMsgAttrs(t, resp.(poolSteps).poolDepth, 3)
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
