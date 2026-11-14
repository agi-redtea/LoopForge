# 自托管构建说明

这个仓库是 [Tencent/LoopForge](https://github.com/Tencent/LoopForge) 的公开 fork，默认按「没有 CodeBuddy 账号、不接腾讯内网」来跑。上游代码都还在，靠下面的开关关掉。想恢复上游行为，设 `AGENTFLOW_SELF_HOSTED=0`，或者只打开需要的那几项。

改动都是加开关和薄封装，没有删上游路径。合 upstream `main` 时优先保留这些新文件，再处理上游文件上的小改动。

## 总开关

| 变量 | 默认 | 作用 |
|---|---|---|
| `AGENTFLOW_SELF_HOSTED` | 开（未设置即开） | `0` / `false` 恢复上游默认：CodeBuddy 可见，腾讯集成可用 |
| `AGENTFLOW_DISABLE_CODEBUDDY` 或 `LOOPFORGE_DISABLE_CODEBUDDY` | 自托管时关 CodeBuddy | `1` 强制隐藏 CodeBuddy，即使 `SELF_HOSTED=0` |
| `AGENTFLOW_ENABLE_CODEBUDDY` 或 `LOOPFORGE_ENABLE_CODEBUDDY` | | `1` 强制打开 CodeBuddy。同时设了 DISABLE 时 DISABLE 优先 |
| `AGENTFLOW_DEFAULT_EXECUTOR` 或 `LOOPFORGE_DEFAULT_EXECUTOR` | `codex`（CodeBuddy 关闭时） | `codex`、`claude-code`（或 `claude`）、`codebuddy`。只改写内置角色上仍然是 `codebuddy` 的执行器。角色自己没填执行器时保持空白，工作流节点可以另外指定；SOP 预览和「两边都没填」的节点才回退到这个默认值 |
| `AGENTFLOW_DEFAULT_MODEL` 或 `LOOPFORGE_DEFAULT_MODEL` | 空 | 回退 CLI 不带 `--model`，用 CLI 自己的默认模型。设了之后，澄清 / 评测 / 摘要会把空的 CodeBuddy「default」换成这个 id |

内置角色（`baseline-align-resolver`、`af-preview-builder`、`loopforge-*`、`generic-*-delivery-*`、`issue-fix-production-*`、`product-*`、`frontend-*`）如果执行器还是 `codebuddy`，读取时改成默认执行器。用户自己建的角色，或者已经改成别的执行器的角色，不动。

`migrations/032_self_hosted_default_executor.sql` 只把 `baseline-align-resolver` 这一行从 `codebuddy` 改成 `codex`，可重复执行。默认执行器是 Claude Code 时不要跑这条，否则这一行会停在 codex 上（读路径只改写仍然是 `codebuddy` 的值）。

## 腾讯集成

每一项都是：`AGENTFLOW_ENABLE_<NAME>=1` 打开，`AGENTFLOW_DISABLE_<NAME>=1` 关掉，否则跟随 `AGENTFLOW_SELF_HOSTED`（自托管默认关）。`LOOPFORGE_ENABLE_*` / `LOOPFORGE_DISABLE_*` 同样生效，DISABLE 优先。

| NAME | 关掉之后 |
|---|---|
| `IOA` | `/api/auth/exchange` 返回 403，不访问 iOA |
| `GONGFENG` | 工蜂 OAuth 状态为未启用，开始授权返回 403；同步不再拉工蜂 issue |
| `TAPD` | 导入 / 重新同步 TAPD 返回 403；同步不再拉 TAPD |
| `SKILLHUB` | `/api/skillhub/*` 返回 403，不会请求市场 |
| `WECOM` | 不能新建或修改企业微信机器人（403），导航里的「通道管理」隐藏 |

`AGENTFLOW_SKILLHUB_URL`（或 `LOOPFORGE_SKILLHUB_URL`）覆盖市场地址，优先于原来的 `SKILLHUB_BASE_URL`。留空时客户端仍用历史上的 `https://api.skillhub.cn`，但自托管默认不会去连。

界面：TAPD / 工蜂入口本来就带 `data-legacy-tracker` 且默认隐藏，只有对应开关打开才显示。SkillHub 市场按钮同理。CodeBuddy 隐藏后，个人凭证里的 CodeBuddy 卡片和「共享 CodeBuddy Key」页签不再出现。

自托管时，澄清和助手的系统提示末尾会加一句：不要假设 TAPD、工蜂、iOA、企业微信或 CodeBuddy。种子 SQL 里的模板正文没有改（旧迁移不回写），这段运行时说明盖住那些假设。

## 执行器、密钥和 Base URL

澄清、运行评测、通知摘要、机器人回复、定时任务和 SOP 预览走同一条助手 CLI 路径（`runAssistantCLI`）。选了 `codex:` / `claude-code:` 就用该用户在个人配置里保存的 API Key 和 Base URL：

- Codex：`CODEX_API_KEY`，`wire_api=responses` 的 Base URL（`CodexBaseURL`）
- Claude Code：`ANTHROPIC_API_KEY`，`ANTHROPIC_BASE_URL`（`ClaudeCodeBaseURL`）

没有 CodeBuddy Key、但配了 Codex 或 Claude 时，不会再出现「需求责任人未配置 CodeBuddy API Key」。两个都没配时，自托管提示配置 Codex 或 Claude Code。

模型 id 用 `default`（或留空）时不向 CLI 传 `--model`，让网关 / CLI 用自己的默认模型。自建网关示例：`http://magpie:3425/v1` 填在对应用户的 Base URL 里。

会话续接：除了原来的 `.codebuddy/projects`，还会看 Claude 的 `.claude/projects` 和 Codex 的 `codex-threads.json`。

## Docker 和遥测

根 `Dockerfile`：

- `LOOPFORGE_INSTALL_CODEBUDDY=0`（默认）不安装 `@tencent-ai/codebuddy-code`。设为 `1` 恢复上游镜像里的那次安装。
- Claude Code 仍是第一次使用时安装（`EnsureClaudeCode`），许可证不允许打进镜像。
- 镜像默认带上 `DISABLE_TELEMETRY=1`、`DO_NOT_TRACK=1`、`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`、`CODEX_DISABLE_TELEMETRY=1`。Codex 官方没有公开的遥测开关，`CODEX_DISABLE_TELEMETRY` 是尽量关掉，不保证 CLI 认这个变量。平台自己的 OTEL 采集不受影响。

自托管时，内置沙箱 Dockerfile 会去掉 `mirrors.cloud.tencent.com` 和 `goproxy.cn`，并把 `CODEBUDDY_PACKAGE` 置空。`AGENTFLOW_SELF_HOSTED=0` 时文本与上游一致。已经写进数据库的旧 Dockerfile 不会自动改写。

## 恢复上游行为

```bash
AGENTFLOW_SELF_HOSTED=0
```

或者只打开某一项，例如 `AGENTFLOW_ENABLE_TAPD=1`、`AGENTFLOW_ENABLE_CODEBUDDY=1`。镜像要带 CodeBuddy 时：

```bash
docker build --build-arg LOOPFORGE_INSTALL_CODEBUDDY=1 .
```

## 往 upstream main 上变基

1. 新文件可以整文件保留：`internal/forkconfig/`、`internal/httpapi/platform_executor.go`、`clarify_executor.go`、`fork_features.go`、`internal/repo/generic_docker_selfhosted.go`、`internal/runtime/isolated_selfhosted.go`、`migrations/032_self_hosted_default_executor.sql`、本文件。
2. 不要改 `migrations/023_role_prompt.sql` 和 `migrations/seed/001_public_catalog.sql`。默认执行器只在读取时替换，外加 032。
3. 上游文件上的改动都是加分支或包一层，冲突时保留上游函数体，再把开关判断加回去。
4. 测试默认用 `AGENTFLOW_SELF_HOSTED=0`（各包 `TestMain`），这样上游断言不用跟着自托管默认走。新测试自己 `t.Setenv("AGENTFLOW_SELF_HOSTED", "1")`。

## 这版没有改掉的地方

- 种子模板正文里仍可能提到 TAPD / 工蜂。运行时提示会补一句自托管说明，正文本身留着，方便以后打开那些集成。
- 源码里的腾讯主机名（`copilot.tencent.com`、`api.skillhub.cn`、工蜂常量）还在，关掉开关后不会去连。
- 企业微信只挡住了保存机器人和导航入口。已经配好的机器人进程代码还在，打开 `WECOM` 就能用。
- 数据库里已经存下来的沙箱 Dockerfile 不会被 032 改写。新环境用自托管构建逻辑；旧行要换镜像的话，在环境设置里更新 Dockerfile，或重新生成内置行。
