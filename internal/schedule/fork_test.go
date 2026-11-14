package schedule

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	os.Exit(m.Run())
}
