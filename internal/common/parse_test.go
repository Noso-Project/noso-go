package common

import (
	"testing"
)

const (
	JOINOK_1 = "JOINOK N6VxgLSpbni8kLbyUAjYXdHCPt2VEp 020000000 PoolData 37873 E1151A4F79E6394F6897A913ADCD476B 11 0 102 0 -30 42270 3"
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
		resp, err := parse(JOINOK_1)

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
