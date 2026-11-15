package forkconfig

import "testing"

func TestSelfHostedDefaults(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "")
	t.Setenv("AGENTFLOW_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "")
	t.Setenv("AGENTFLOW_ENABLE_TAPD", "")
	t.Setenv("AGENTFLOW_DISABLE_TAPD", "")
	t.Setenv("AGENTFLOW_DISABLE_WECOM", "")
	t.Setenv("LOOPFORGE_DISABLE_WECOM", "")
	if !SelfHosted() || CodeBuddyEnabled() || Feature(FeatureTAPD) || Feature(FeatureIOA) || Feature(FeatureGongfeng) || Feature(FeatureSkillHub) || !Feature(FeatureWeCom) {
		t.Fatalf("self-hosted defaults: hosted=%v codebuddy=%v tapd=%v ioa=%v gongfeng=%v skillhub=%v wecom=%v", SelfHosted(), CodeBuddyEnabled(), Feature(FeatureTAPD), Feature(FeatureIOA), Feature(FeatureGongfeng), Feature(FeatureSkillHub), Feature(FeatureWeCom))
	}
	if DefaultExecutor() != "codex" || DefaultEngine() != "codex" {
		t.Fatalf("default executor = %s / %s, want codex", DefaultExecutor(), DefaultEngine())
	}
}

func TestUpstreamProfileRestoresCodeBuddyAndTencent(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	t.Setenv("AGENTFLOW_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "")
	t.Setenv("AGENTFLOW_DISABLE_GONGFENG", "")
	if SelfHosted() || !CodeBuddyEnabled() || !Feature(FeatureGongfeng) || !Feature(FeatureWeCom) {
		t.Fatalf("upstream profile: hosted=%v codebuddy=%v gongfeng=%v wecom=%v", SelfHosted(), CodeBuddyEnabled(), Feature(FeatureGongfeng), Feature(FeatureWeCom))
	}
	if DefaultExecutor() != "codebuddy" {
		t.Fatalf("upstream default executor = %s", DefaultExecutor())
	}
}

func TestExplicitSwitchesOverrideProfile(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "1")
	t.Setenv("AGENTFLOW_DISABLE_SKILLHUB", "1")
	t.Setenv("AGENTFLOW_ENABLE_TAPD", "1")
	if CodeBuddyEnabled() {
		t.Fatal("LOOPFORGE_DISABLE_CODEBUDDY should hide CodeBuddy even upstream")
	}
	if Feature(FeatureSkillHub) || !Feature(FeatureTAPD) {
		t.Fatalf("skillhub=%v tapd=%v", Feature(FeatureSkillHub), Feature(FeatureTAPD))
	}
	t.Setenv("AGENTFLOW_SELF_HOSTED", "1")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "1")
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "claude-code")
	if !CodeBuddyEnabled() || DefaultExecutor() != "claude" || DefaultEngine() != "claude-code" {
		t.Fatalf("enable codebuddy with claude default: enabled=%v executor=%s engine=%s", CodeBuddyEnabled(), DefaultExecutor(), DefaultEngine())
	}
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_WECOM", "1")
	if Feature(FeatureWeCom) {
		t.Fatal("LOOPFORGE_DISABLE_WECOM should turn WeCom off in the self-hosted profile")
	}
}

func TestMapStockExecutorLeavesCustomRoles(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "1")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DISABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DEFAULT_EXECUTOR", "")
	if got := MapStockExecutor("baseline-align-resolver", "codebuddy"); got != "codex" {
		t.Fatalf("stock role = %s, want codex", got)
	}
	if got := MapStockExecutor("loopforge-implement", "codebuddy"); got != "codex" {
		t.Fatalf("template role = %s, want codex", got)
	}
	if got := MapStockExecutor("my-reviewer", "codebuddy"); got != "codebuddy" {
		t.Fatalf("custom role = %s, want codebuddy left alone", got)
	}
	if got := MapStockExecutor("developer", ""); got != "" {
		t.Fatalf("empty custom executor = %q, want empty so the node can supply one", got)
	}
	if got := MapStockExecutor("loopforge-implement", "claude"); got != "claude" {
		t.Fatalf("customized template executor = %s", got)
	}
	if got := MapPreviewExecutor(""); got != "codex" {
		t.Fatalf("preview executor = %s, want codex", got)
	}
	t.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	if got := MapStockExecutor("baseline-align-resolver", "codebuddy"); got != "codebuddy" {
		t.Fatalf("upstream stock role = %s", got)
	}
}

func TestSkillHubURLOverride(t *testing.T) {
	t.Setenv("SKILLHUB_BASE_URL", "https://api.skillhub.cn")
	t.Setenv("AGENTFLOW_SKILLHUB_URL", "http://market.internal/v1/")
	if got := SkillHubURL(); got != "http://market.internal/v1" {
		t.Fatalf("skillhub url = %s", got)
	}
}
