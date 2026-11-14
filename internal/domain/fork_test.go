package domain

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	os.Exit(m.Run())
}

func TestPreviewExecutorFollowsSelfHostedDefault(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "1")
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DISABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	c := &PreviewContract{Enabled: true, Model: "default"}
	if got := c.ExecutorOrCodeBuddy(); got != "codex" {
		t.Fatalf("executor = %s, want codex", got)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "claude-code")
	fresh := &PreviewContract{Enabled: true, Model: "default"}
	if got := fresh.ExecutorOrCodeBuddy(); got != "claude" {
		t.Fatalf("executor = %s, want claude", got)
	}
}
