package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"agentflow/internal/engine"
	"agentflow/internal/repo"
	"agentflow/internal/runtime"
	"agentflow/internal/tracker"
)

func TestSelfHostedBootstrapHidesCodeBuddyAndTencentFeatures(t *testing.T) {
	t.Setenv("AGENTFLOW_SELF_HOSTED", "1")
	t.Setenv("AGENTFLOW_ENABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_DISABLE_CODEBUDDY", "")
	t.Setenv("LOOPFORGE_DISABLE_CODEBUDDY", "")
	t.Setenv("AGENTFLOW_ENABLE_IOA", "")
	t.Setenv("AGENTFLOW_ENABLE_GONGFENG", "")
	t.Setenv("AGENTFLOW_ENABLE_TAPD", "")
	t.Setenv("AGENTFLOW_ENABLE_SKILLHUB", "")
	t.Setenv("AGENTFLOW_ENABLE_WECOM", "")
	t.Setenv("AGENTFLOW_DISABLE_WECOM", "")
	t.Setenv("LOOPFORGE_DISABLE_WECOM", "")

	root := t.TempDir()
	s, err := repo.OpenTest()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	h := New(s, engine.New(s, runtime.NewManager(filepath.Join(root, "workspaces"))), tracker.NewService(), "", "")

	bootRec := httptest.NewRecorder()
	h.ServeHTTP(bootRec, httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil))
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap = %d %s", bootRec.Code, bootRec.Body.String())
	}
	var boot struct {
		Executors []struct {
			ID string `json:"id"`
		} `json:"executors"`
		Features map[string]any `json:"features"`
	}
	if err := json.Unmarshal(bootRec.Body.Bytes(), &boot); err != nil {
		t.Fatal(err)
	}
	if boot.Features["codebuddy"] != false || boot.Features["tapd"] != false || boot.Features["gongfeng"] != false || boot.Features["skillhub"] != false || boot.Features["ioa"] != false || boot.Features["wecom"] != true || boot.Features["selfHosted"] != true {
		t.Fatalf("features = %#v", boot.Features)
	}
	if boot.Features["defaultExecutor"] != "codex" {
		t.Fatalf("default executor = %#v", boot.Features["defaultExecutor"])
	}
	for _, ex := range boot.Executors {
		if ex.ID == "codebuddy" {
			t.Fatal("bootstrap still lists codebuddy")
		}
	}

	for _, path := range []string{"/api/auth/exchange?code=abc", "/api/skillhub/skills", "/api/issues/import-tapd"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if path == "/api/issues/import-tapd" {
			req = httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"projectId":"platform","url":"https://tapd.cn/1"}`))
			req.Header.Set("Content-Type", "application/json")
		}
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s = %d %s", path, rec.Code, rec.Body.String())
		}
	}

	gf := httptest.NewRecorder()
	h.ServeHTTP(gf, httptest.NewRequest(http.MethodGet, "/api/auth/gongfeng", nil))
	if gf.Code != http.StatusOK || !strings.Contains(gf.Body.String(), `"enabled":false`) {
		t.Fatalf("gongfeng status = %d %s", gf.Code, gf.Body.String())
	}
	start := httptest.NewRecorder()
	h.ServeHTTP(start, httptest.NewRequest(http.MethodGet, "/api/auth/gongfeng/start", nil))
	if start.Code != http.StatusForbidden {
		t.Fatalf("gongfeng start = %d %s", start.Code, start.Body.String())
	}

	bot := httptest.NewRecorder()
	botReq := httptest.NewRequest(http.MethodPost, "/api/notification-bots", strings.NewReader(`{"projectId":"platform","name":"n","botId":"b","secret":"s","enabled":true}`))
	botReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(bot, botReq)
	if bot.Code == http.StatusForbidden || strings.Contains(bot.Body.String(), "当前部署已关闭企业微信通知") {
		t.Fatalf("wecom stays available in the self-hosted profile: %d %s", bot.Code, bot.Body.String())
	}

	t.Setenv("AGENTFLOW_DISABLE_WECOM", "1")
	closed := httptest.NewRecorder()
	closedReq := httptest.NewRequest(http.MethodPost, "/api/notification-bots", strings.NewReader(`{"projectId":"platform","name":"n","botId":"b","secret":"s","enabled":true}`))
	closedReq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(closed, closedReq)
	if closed.Code != http.StatusForbidden || !strings.Contains(closed.Body.String(), "当前部署已关闭企业微信通知") {
		t.Fatalf("disabled wecom = %d %s", closed.Code, closed.Body.String())
	}
}
