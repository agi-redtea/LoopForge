package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"agentflow/internal/domain"
)

// runClarifyTurnOnEngine runs one clarification turn on Codex or Claude Code
// through the same CLI path the assistant chat uses, so the user's Base URL
// and API key are applied and the stream is translated back to the events
// the clarify UI already renders.
func (h *streamHandlers) runClarifyTurnOnEngine(ctx context.Context, workDir, model, prompt, sessionID, homeDir string, resume bool, user domain.User, emit func(event string, payload any), opts clarifyTurnOpts) (any, string, string, error) {
	root := filepath.Dir(homeDir)
	if strings.TrimSpace(root) == "" || root == "." {
		root = homeDir
	}
	out, sid, err := runAssistantCLI(ctx, assistantCLIRequest{
		Prompt:       prompt,
		Model:        model,
		SessionID:    sessionID,
		Resume:       resume,
		SessionRoot:  root,
		WorkDir:      workDir,
		UserID:       user.ID,
		Native:       true,
		SubmitOnly:   opts.SubmitOnly,
		MCPConfig:    opts.MCPConfig,
		Settings:     opts.Settings,
		SystemPrompt: "你是研发工作项分析师。完成后调用 submit_clarification。",
	}, user, func(ev assistantStreamEvent) {
		if emit == nil {
			return
		}
		text := strings.TrimSpace(ev.Text)
		if text == "" {
			return
		}
		kind := ev.Kind
		if kind == "" {
			kind = "text"
		}
		emit("thought", map[string]string{"kind": kind, "text": text})
	})
	if sid == "" {
		sid = sessionID
	}
	if err != nil {
		return nil, out, sid, err
	}
	return nil, strings.TrimSpace(out), sid, nil
}

func clarifySessionFileExists(pattern string) bool {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	for _, file := range files {
		info, err := os.Stat(file)
		if err == nil && info.Mode().IsRegular() && info.Size() > 0 {
			return true
		}
	}
	return false
}
