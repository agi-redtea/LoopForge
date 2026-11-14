package httpapi

import (
	"errors"
	"fmt"
	"strings"

	"agentflow/internal/domain"
	"agentflow/internal/forkconfig"
	"agentflow/internal/runtime"
)

// resolvePlatformModel picks the CLI model for clarify, evals, reports, bots
// and other entry points that upstream hard-wires to CodeBuddy.
//
// An explicit "codex:" or "claude-code:" id keeps that CLI and its per-user
// API key and Base URL. An unprefixed CodeBuddy id stays on CodeBuddy when
// that CLI is enabled and the user has a key; otherwise it moves to the
// configured fallback CLI (codex by default, or claude-code) when that user
// has a key. The CodeBuddy key error is returned only when CodeBuddy is the
// selected CLI and no other key is configured.
func resolvePlatformModel(user domain.User, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	engine, model := assistantModelEngine(requested)
	if engine != assistantEngineCodeBuddy {
		if model != "" && !assistantValidEngineModel(model) {
			return "", fmt.Errorf("模型 ID 无效")
		}
		if key, _ := assistantEngineCredential(user, engine); key == "" {
			return "", errors.New(assistantEngineMissingKey(engine))
		}
		if model == "" {
			model = "default"
		}
		return engine + ":" + model, nil
	}
	if forkconfig.CodeBuddyEnabled() {
		if key, _ := assistantEngineCredential(user, assistantEngineCodeBuddy); key != "" {
			return assistantAllowedModel(requested, assistantKnownCatalog()), nil
		}
	}
	if alt, bare, ok := fallbackPlatformEngine(user, requested); ok {
		return alt + ":" + bare, nil
	}
	if !forkconfig.CodeBuddyEnabled() {
		return "", errors.New("请先在个人配置中配置 Codex 或 Claude Code API Key")
	}
	return "", errors.New("需求责任人未配置 CodeBuddy API Key，请在个人配置中配置")
}

func fallbackPlatformEngine(user domain.User, requested string) (engine, model string, ok bool) {
	bare := strings.TrimSpace(requested)
	if bare == "" || bare == "default" || !assistantValidEngineModel(bare) {
		bare = "default"
		if custom := forkconfig.DefaultModel(); custom != "" && assistantValidEngineModel(custom) {
			bare = custom
		}
	}
	seen := map[string]bool{}
	for _, engine := range []string{forkconfig.DefaultEngine(), assistantEngineCodex, assistantEngineClaude} {
		if engine == "" || engine == assistantEngineCodeBuddy || seen[engine] {
			continue
		}
		seen[engine] = true
		if key, _ := assistantEngineCredential(user, engine); key != "" {
			return engine, bare, true
		}
	}
	return "", "", false
}

// supportedPlatformEngine reports whether runAssistantCLI may run a
// non-CodeBuddy call that upstream limited to CodeBuddy (reports, evals,
// clarification). Unknown engines still return errAssistantEngineUnsupported.
func supportedPlatformEngine(engine string) bool {
	switch engine {
	case assistantEngineCodex, assistantEngineClaude:
		return true
	default:
		return runtime.IsAgentCLI(engine)
	}
}

func withSelfHostedPromptNote(prompt string) string {
	note := forkconfig.PromptNote()
	if note == "" {
		return prompt
	}
	return prompt + "\n\n" + note
}

func redactExecutorSecrets(text string, user domain.User) string {
	for _, secret := range []string{user.CodeBuddyAPIKey, user.CodexAPIKey, user.ClaudeCodeAPIKey} {
		secret = strings.TrimSpace(secret)
		if secret != "" {
			text = strings.ReplaceAll(text, secret, "[redacted]")
		}
	}
	return text
}
