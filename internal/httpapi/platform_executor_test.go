package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentflow/internal/domain"
)

func TestResolvePlatformModelUsesConfiguredCLI(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "1")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DISABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "")
	t.Setenv("AGENTFLOW_DEFAULT_MODEL", "")

	codexUser := domain.User{CodexAPIKey: "sk-codex", CodexBaseURL: "http://magpie:3425/v1"}
	got, err := resolvePlatformModel(codexUser, "")
	if err != nil || got != "codex:default" {
		t.Fatalf("fallback = %q %v, want codex:default", got, err)
	}
	if _, err := resolvePlatformModel(codexUser, "glm-5.1"); err != nil {
		t.Fatal(err)
	} else if got, _ := resolvePlatformModel(codexUser, "glm-5.1"); got != "codex:glm-5.1" {
		t.Fatalf("bare model = %s", got)
	}
	got, err = resolvePlatformModel(codexUser, "claude-code:sonnet")
	if err == nil || got != "" {
		t.Fatalf("missing claude key = %q %v", got, err)
	}
	if !strings.Contains(err.Error(), "Claude") {
		t.Fatalf("error = %v", err)
	}

	claudeUser := domain.User{ClaudeCodeAPIKey: "sk-ant", ClaudeCodeBaseURL: "http://magpie:3425"}
	got, err = resolvePlatformModel(claudeUser, "default")
	if err != nil || got != "claude-code:default" {
		t.Fatalf("claude fallback = %q %v", got, err)
	}
	_, err = resolvePlatformModel(domain.User{}, "default")
	if err == nil || strings.Contains(err.Error(), "CodeBuddy API Key") {
		t.Fatalf("hidden codebuddy should not ask for a CodeBuddy key: %v", err)
	}
}

func TestResolvePlatformModelKeepsCodeBuddyWhenEnabled(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	user := domain.User{CodeBuddyAPIKey: "ck_test"}
	got, err := resolvePlatformModel(user, "glm-5.1")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" || len(got) > 4 && got[:4] == "codex" {
		t.Fatalf("upstream model = %s", got)
	}
	_, err = resolvePlatformModel(domain.User{}, "")
	if err == nil || !strings.Contains(err.Error(), "CodeBuddy API Key") {
		t.Fatalf("upstream missing key = %v", err)
	}
}

func TestCanResumeClarifySessionSeesClaudeAndCodex(t *testing.T) {
	home := t.TempDir()
	session := "sess-1"
	if canResumeClarifySession(home, session, 0, 0) {
		t.Fatal("empty home resumed")
	}
	claudeDir := filepath.Join(home, ".claude", "projects", "repo")
	if err := os.MkdirAll(claudeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	claudeID := assistantClaudeSessionID(session)
	if err := os.WriteFile(filepath.Join(claudeDir, claudeID+".jsonl"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !canResumeClarifySession(home, session, 0, 0) {
		t.Fatal("claude session was not resumable")
	}
	root := t.TempDir()
	home = filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeCodexThread(root, "thread-session", "thread_123"); err != nil {
		t.Fatal(err)
	}
	if !canResumeClarifySession(home, "thread-session", 0, 0) {
		t.Fatal("codex thread was not resumable")
	}
}
