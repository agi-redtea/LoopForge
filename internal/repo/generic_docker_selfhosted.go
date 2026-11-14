package repo

import (
	"strings"

	"agentflow/internal/forkconfig"
)

// selfHostedDockerfile drops the Tencent Go mirror and the CodeBuddy package
// from the built-in sandbox Dockerfile. The original text stays in
// genericBackendDockerfileContent so upstream merges stay a straight diff.
func selfHostedDockerfile(content string) string {
	if !forkconfig.SelfHosted() {
		return content
	}
	content = strings.ReplaceAll(content,
		"      \"https://mirrors.cloud.tencent.com/go/go${GO_VERSION}.linux-${goarch}.tar.gz\" \\\n",
		"")
	content = strings.ReplaceAll(content, `GOPROXY="https://goproxy.cn,direct"`, `GOPROXY="https://proxy.golang.org,direct"`)
	content = strings.ReplaceAll(content, "GOSUMDB=sum.golang.google.cn", "GOSUMDB=sum.golang.org")
	content = strings.Replace(content, "ARG CODEBUDDY_PACKAGE=@tencent-ai/codebuddy-code", "ARG CODEBUDDY_PACKAGE=", 1)
	return content
}
