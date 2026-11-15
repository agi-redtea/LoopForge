package forkconfig

import "strings"

// StockRole reports whether agentID is a platform catalog role that ships
// with executorId codebuddy. User-created roles are not stock, so a custom
// role left on CodeBuddy is not rewritten.
func StockRole(agentID string) bool {
	id := strings.TrimSpace(agentID)
	switch id {
	case "", "baseline-align-resolver", "af-preview-builder":
		// An empty id is the platform fallback (no role chosen yet), which
		// upstream fills with CodeBuddy. Callers that already know a custom
		// role pass that role's id.
		return true
	}
	for _, prefix := range []string{
		"loopforge-",
		"generic-backend-delivery-",
		"generic-frontend-delivery-",
		"generic-product-delivery-",
		"issue-fix-production-",
		"product-",
		"frontend-",
	} {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}
