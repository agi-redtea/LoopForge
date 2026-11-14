-- Idempotent: the stock baseline-align role still pinned to CodeBuddy moves
-- to codex. Other stock roles are remapped when they are read
-- (forkconfig.MapStockExecutor), so a role whose executor was edited is left
-- alone and this script does not touch generic template blobs.
--
-- Installs that want Claude Code as the default should set
-- AGENTFLOW_DEFAULT_EXECUTOR=claude-code and skip this script. A row already
-- stored as "codex" is not rewritten to claude-code; only a stored
-- "codebuddy" id is remapped at read time.

UPDATE agents
SET payload = JSON_SET(payload, '$.executorId', 'codex')
WHERE id = 'baseline-align-resolver'
  AND JSON_UNQUOTE(JSON_EXTRACT(payload, '$.executorId')) = 'codebuddy';
