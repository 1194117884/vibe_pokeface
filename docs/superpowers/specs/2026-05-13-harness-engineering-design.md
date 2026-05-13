# Harness Engineering 改造方案

## 概述

将 vibe_pokeface 项目从"描述性 CLAUDE.md + 无验证 + 零前端测试"状态，升级为完整 Harness Engineering 体系：约束系统、自动验证、长期记忆、任务规范、前端测试。

## 工作包划分

4 个独立工作包，可完全并行执行（无文件冲突）。

---

## 工作包 A: CLAUDE.md 重写

**目标**：从项目说明书升级为 AI 行为约束系统

**变更文件**：`CLAUDE.md`

### 新增章节

| 章节 | 内容 |
|---|---|
| `## Workflow` | 先输出计划 → 等待确认 → 小步执行 → 每步验证 |
| `## Coding Rules` | 禁止 any、禁止大规模重构、必须 type-safe、Server 用 Go、Frontend 用 TypeScript |
| `## Verification` | 修改后必须执行 `go vet`、`npm run lint`、`npx tsc --noEmit`、`go test`（相关包） |
| `## Scope Constraints` | 禁止修改：数据库迁移、Docker 配置、LiveKit 配置。Server 只改 internal/，Frontend 只改 app/ components/ lib/ |
| `## Task Granularity` | 一次任务只做一件事，超过 3 个文件必须拆分子任务，完成一步汇报一次 |

### 保留内容

- 现有架构描述（Project Overview、Server Package Layout、Frontend Structure、Game Flow、Key Infrastructure）
- Commands 部分

---

## 工作包 B: Settings + Hooks

**目标**：添加自动验证机制 + 清理配置

**变更文件**：`.claude/settings.local.json`

### Post-Tool Hooks

每次编辑/写入文件后自动执行：

```json
{
  "hooks": {
    "post-tool": {
      "edit": [
        "cd server && go vet ./..."
      ],
      "write": [
        "cd server && go vet ./..."
      ]
    }
  }
}
```

前端验证（lint + typecheck）因可能耗时较长，作为手动要求而非 hook，避免阻塞编辑流程。

### Settings 清理

- 修复 `vibe_pokefake` 拼写错误路径
- 移除 `Bash(cat)` 过度宽松权限
- 补充前端相关权限（`npx vitest`、`npx tsc --noEmit`）

---

## 工作包 C: 长期记忆体系

**目标**：建立持久化上下文，防止跨 session 失忆

**变更文件**：`docs/architecture.md`、`docs/decisions.md`、`docs/api-rules.md`

### docs/architecture.md

记录架构决策：

- 为什么用 chi router（轻量、兼容 net/http）
- 为什么用 sqlx（轻量、直接 SQL 控制）
- 为什么用 gorilla/websocket（成熟稳定）
- 为什么用 JWT（无状态、适合 WebSocket 场景）
- 为什么用 LiveKit（开源 WebRTC SFU）
- 前后端通信方式（REST + WebSocket 混合）
- 游戏引擎设计（GameEngine 接口 + doudizhu 实现）

### docs/decisions.md

ADR 格式：

- ADR-001: SQLite → MySQL 迁移
- ADR-002: Card ID 编码方案（0-53）
- ADR-003: AI 玩家采用 LLM 而非规则引擎
- ADR-004: Starbucks 设计系统

### docs/api-rules.md

API 约定：

- REST 端点命名（`/api/{resource}`）
- 错误响应格式（`{error: string, code: int}`）
- 分页格式（`{data: [], total: int, page: int, pageSize: int}`）
- WebSocket 消息格式（`{type: string, payload: object}`）
- JWT 认证方式（Authorization: Bearer header / ws query param）

---

## 工作包 D: 前端测试初始化

**目标**：建立前端测试基础，覆盖核心通信模块

**变更文件**：
- `frontend/package.json`（添加 vitest 依赖 + test script）
- `frontend/vitest.config.ts`（新建）
- `frontend/lib/__tests__/ws-game.test.ts`（新建）
- `frontend/lib/__tests__/api-client.test.ts`（新建）

### 技术选型

- Vitest（与 Vite/Next.js 生态兼容、速度快）
- @testing-library/react（组件测试，可选）
- msw（API mock，可选，初期直接用 mock class）

### 测试计划

| 模块 | 测试内容 |
|---|---|
| `lib/api-client.ts` | token 存储/读取、login/register 请求构造、token 过期处理、错误响应解析 |
| `lib/ws-game.ts` | WebSocket 连接建立、消息发送/接收、重连机制、房间操作（join/leave/action） |
| `lib/livekit-client.ts` | token 获取、连接/断开（基础） |

### 测试策略

- 网络层用 mock class 替代真实 WebSocket/fetch
- 不依赖真实后端
- 每次 commit 前可运行 `npm test`

---

## 执行顺序

```
Step 1: 并行执行包 A + B + C（无依赖，互不冲突）
Step 2: 执行包 D（可在 Step 1 完成后单独启动）
```

所有包可在同一轮并行启动。
