package runtime

import (
	"strings"

	"agentflow/internal/forkconfig"
)

// effectiveIsolatedDockerfile is the sandbox image the self-hosted profile
// builds. The published constant still names codebuddy so existing equality
// checks and upstream images stay intact.
func effectiveIsolatedDockerfile() string {
	if !forkconfig.SelfHosted() {
		return IsolatedDockerfile
	}
	return strings.Replace(IsolatedDockerfile, " @tencent-ai/codebuddy-code", "", 1)
}
