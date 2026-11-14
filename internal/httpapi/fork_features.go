package httpapi

import (
	"agentflow/internal/domain"
	"agentflow/internal/forkconfig"
)

func forkFeaturePayload() map[string]any {
	return map[string]any{
		"selfHosted":      forkconfig.SelfHosted(),
		"codebuddy":       forkconfig.CodeBuddyEnabled(),
		"defaultExecutor": forkconfig.DefaultExecutor(),
		"defaultModel":    forkconfig.DefaultModel(),
		"ioa":             forkconfig.Feature(forkconfig.FeatureIOA),
		"gongfeng":        forkconfig.Feature(forkconfig.FeatureGongfeng),
		"tapd":            forkconfig.Feature(forkconfig.FeatureTAPD),
		"skillhub":        forkconfig.Feature(forkconfig.FeatureSkillHub),
		"wecom":           forkconfig.Feature(forkconfig.FeatureWeCom),
	}
}

func withoutCodeBuddyExecutors(in []domain.ExecutorDefinition) []domain.ExecutorDefinition {
	if forkconfig.CodeBuddyEnabled() {
		return in
	}
	out := make([]domain.ExecutorDefinition, 0, len(in))
	for _, ex := range in {
		if ex.ID == "codebuddy" || ex.Kind == "codebuddy" {
			continue
		}
		out = append(out, ex)
	}
	return out
}

func hideCodeBuddyCatalog(catalog assistantModelCatalog) assistantModelCatalog {
	if forkconfig.CodeBuddyEnabled() {
		return catalog
	}
	kept := make([]assistantModel, 0, len(catalog.Models))
	for _, model := range catalog.Models {
		if model.Engine == "" || model.Engine == assistantEngineCodeBuddy {
			continue
		}
		kept = append(kept, model)
	}
	catalog.Models = kept
	catalog.SponsoredModel = ""
	catalog.SponsoredModels = nil
	if len(kept) > 0 {
		catalog.Default = kept[0].ID
	}
	return catalog
}
