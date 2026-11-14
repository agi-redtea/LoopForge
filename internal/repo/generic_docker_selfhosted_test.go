package repo

import (
	"strings"
	"testing"
)

func TestSelfHostedSandboxDockerfileDropsTencentMirror(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "1")
	content := GenericBackendDockerfile().Content
	if strings.Contains(content, "mirrors.cloud.tencent.com") || strings.Contains(content, "goproxy.cn") {
		t.Fatalf("self-hosted dockerfile still uses a Tencent mirror:\n%s", content)
	}
	if strings.Contains(content, "ARG CODEBUDDY_PACKAGE=@tencent-ai/codebuddy-code") {
		t.Fatal("self-hosted dockerfile still installs codebuddy by default")
	}
	if !strings.Contains(content, "https://dl.google.com/go/") {
		t.Fatal("official Go download URL was removed")
	}
	t.Setenv("AGENTFLOW_SELF_HOSTED", "0")
	upstream := GenericBackendDockerfile().Content
	if !strings.Contains(upstream, "mirrors.cloud.tencent.com") {
		t.Fatal("upstream dockerfile should keep the historical mirror list")
	}
}
