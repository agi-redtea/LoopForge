// Package forkconfig holds the self-hosted deployment switches.
//
// Upstream LoopForge assumes CodeBuddy and several Tencent-internal
// integrations. This fork keeps those code paths and turns them off unless
// an operator opts back in. Defaults match a public self-hosted install.
// Set AGENTFLOW_SELF_HOSTED=0 to restore the upstream defaults in one step.
package forkconfig

import (
	"os"
	"strings"
)

const (
	// FeatureIOA is Tencent iOA / enterprise SSO login.
	FeatureIOA = "ioa"
	// FeatureGongfeng is Tencent Git (工蜂) OAuth and issue sync.
	FeatureGongfeng = "gongfeng"
	// FeatureTAPD is TAPD import and story sync.
	FeatureTAPD = "tapd"
	// FeatureSkillHub is the SkillHub market (api.skillhub.cn by default).
	FeatureSkillHub = "skillhub"
	// FeatureWeCom is the Enterprise WeChat (企业微信) notifier.
	FeatureWeCom = "wecom"
)

// SelfHosted reports whether this process uses the public self-hosted profile.
// The default is on. AGENTFLOW_SELF_HOSTED=0 or false restores upstream
// defaults (CodeBuddy visible, Tencent integrations available).
func SelfHosted() bool {
	return !envFalse("AGENTFLOW_SELF_HOSTED")
}

// CodeBuddyEnabled reports whether CodeBuddy is offered in selectors,
// credential settings, and executor fallbacks.
//
//   - AGENTFLOW_DISABLE_CODEBUDDY=1 or LOOPFORGE_DISABLE_CODEBUDDY=1 forces it off
//   - AGENTFLOW_ENABLE_CODEBUDDY=1 forces it on
//   - otherwise it is off in the self-hosted profile and on upstream
func CodeBuddyEnabled() bool {
	if envTrue("AGENTFLOW_DISABLE_CODEBUDDY") || envTrue("LOOPFORGE_DISABLE_CODEBUDDY") {
		return false
	}
	if envTrue("AGENTFLOW_ENABLE_CODEBUDDY") || envTrue("LOOPFORGE_ENABLE_CODEBUDDY") {
		return true
	}
	return !SelfHosted()
}

// DefaultExecutor is the catalog executor id used when a stock role still
// says codebuddy and CodeBuddy is hidden. AGENTFLOW_DEFAULT_EXECUTOR accepts
// codex (default), claude-code / claude, or codebuddy.
func DefaultExecutor() string {
	switch normalizeExecutor(firstEnv("AGENTFLOW_DEFAULT_EXECUTOR", "LOOPFORGE_DEFAULT_EXECUTOR")) {
	case "claude":
		return "claude"
	case "codebuddy":
		if CodeBuddyEnabled() {
			return "codebuddy"
		}
		return "codex"
	case "codex":
		return "codex"
	default:
		if CodeBuddyEnabled() {
			return "codebuddy"
		}
		return "codex"
	}
}

// DefaultEngine is the model-id prefix for DefaultExecutor
// ("claude-code" for the Claude catalog id "claude").
func DefaultEngine() string {
	if DefaultExecutor() == "claude" {
		return "claude-code"
	}
	return DefaultExecutor()
}

// DefaultModel is an optional model id passed to the fallback CLI when a
// saved selector still says the CodeBuddy "default" model.
// AGENTFLOW_DEFAULT_MODEL empty means the CLI's own default.
func DefaultModel() string {
	return strings.TrimSpace(firstEnv("AGENTFLOW_DEFAULT_MODEL", "LOOPFORGE_DEFAULT_MODEL"))
}

// Feature reports whether a Tencent integration is available.
// AGENTFLOW_ENABLE_<NAME>=1 forces it on, AGENTFLOW_DISABLE_<NAME>=1 forces
// it off. Otherwise iOA, Gongfeng, TAPD and SkillHub follow the self-hosted
// profile (off). Enterprise WeChat stays on in that profile; turn it off
// with AGENTFLOW_DISABLE_WECOM=1 (LOOPFORGE_DISABLE_WECOM works too).
func Feature(name string) bool {
	key := featureKey(name)
	if key == "" {
		return false
	}
	if envTrue("AGENTFLOW_DISABLE_"+key) || envTrue("LOOPFORGE_DISABLE_"+key) {
		return false
	}
	if envTrue("AGENTFLOW_ENABLE_"+key) || envTrue("LOOPFORGE_ENABLE_"+key) {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(name), FeatureWeCom) {
		return true
	}
	return !SelfHosted()
}

// SkillHubURL is the market base URL. AGENTFLOW_SKILLHUB_URL and
// LOOPFORGE_SKILLHUB_URL override SKILLHUB_BASE_URL. An empty result means
// the historical default (https://api.skillhub.cn) inside the client.
func SkillHubURL() string {
	return strings.TrimRight(strings.TrimSpace(firstEnv("AGENTFLOW_SKILLHUB_URL", "LOOPFORGE_SKILLHUB_URL", "SKILLHUB_BASE_URL")), "/")
}

// PromptNote is appended to default assistant prompts when the self-hosted
// profile is on, so stock wording does not assume Tencent-internal systems.
func PromptNote() string {
	if !SelfHosted() {
		return ""
	}
	return "自托管环境：不要假设 TAPD、工蜂、iOA 或 CodeBuddy 可用。企业微信通知可以配置。代码仓库按用户配置的 Git（通常是 GitHub）处理；模型请求走该用户在个人配置里保存的 Base URL 和 API Key。"
}

// MapStockExecutor rewrites the upstream CodeBuddy default on stock catalog
// roles. An empty executor is left empty so a workflow node can still supply
// one. A role whose executor was changed to anything else is left alone,
// including a custom role that still names CodeBuddy on purpose.
func MapStockExecutor(agentID, executorID string) string {
	executorID = strings.TrimSpace(executorID)
	if !strings.EqualFold(executorID, "codebuddy") {
		return executorID
	}
	if CodeBuddyEnabled() || !StockRole(agentID) {
		return executorID
	}
	return DefaultExecutor()
}

// MapPreviewExecutor is MapStockExecutor for the SOP preview build node,
// which upstream hard-wires to CodeBuddy. An empty id is that default.
func MapPreviewExecutor(executorID string) string {
	if strings.TrimSpace(executorID) == "" {
		executorID = "codebuddy"
	}
	return MapStockExecutor("af-preview-builder", executorID)
}

func featureKey(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case FeatureIOA:
		return "IOA"
	case FeatureGongfeng:
		return "GONGFENG"
	case FeatureTAPD:
		return "TAPD"
	case FeatureSkillHub:
		return "SKILLHUB"
	case FeatureWeCom:
		return "WECOM"
	default:
		return ""
	}
}

func normalizeExecutor(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "codex":
		return "codex"
	case "claude", "claude-code":
		return "claude"
	case "codebuddy":
		return "codebuddy"
	default:
		return ""
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func envTrue(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// envFalse reports whether key is explicitly disabled. An unset key is not false,
// so SelfHosted stays on unless the operator opts out.
func envFalse(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "0", "false", "no", "off":
		return true
	default:
		return false
	}
}
