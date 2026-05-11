# Vibe Pokeface — 整体设计方案

## 项目概述

家庭扑克娱乐游戏平台，支持远程在线对战、AI大模型 Bot 补位、语音聊天、表情包互动。拒绝赌博。

## 技术栈

| 层 | 技术 | 部署 |
|---|------|------|
| 前端 | React + Next.js | Vercel |
| 后端 | Go (REST + WebSocket) | VPS |
| 数据库 | MySQL | VPS |
| 实时语音 | LiveKit (自部署) | VPS |
| AI Bot | LLM (GPT-4o/Claude/... 可切换) | 通过 API 调用 |

## 架构方案

方案 A — Go 统一后端同时提供 REST API（认证、房间管理、用户资料）和 WebSocket（实时游戏），Next.js 直接调用 Go 接口。

```
Vercel (Next.js)          VPS
┌─────────────────┐     ┌──────────────────────┐
│  Frontend App    │────▶│  Go Server            │
│  (main + admin)  │     │  ├── REST API         │
│                  │◀────│  ├── WebSocket Hub    │
│  LiveKit Client  │     │  ├── Game Engine      │
│                  │────▶│  ├── AI Bot Module    │
│                  │     │  └── LiveKit Server   │
└─────────────────┘     ├──────────────────────┤
                         │  MySQL                │
                         └──────────────────────┘
```

## 项目目录结构

```
vibe_pokeface/
├── frontend/
│   ├── app/
│   │   ├── (main)/           # 用户主界面
│   │   │   ├── lobby/        # 大厅
│   │   │   └── room/[id]/    # 游戏房间
│   │   ├── admin/            # CMS 管理端
│   │   │   ├── dashboard/
│   │   │   ├── users/
│   │   │   ├── rooms/
│   │   │   ├── llm-config/
│   │   │   ├── ai-characters/
│   │   │   └── stats/
│   │   ├── auth/
│   │   └── api/              # BFF 代理
│   ├── lib/
│   │   ├── ws-client.ts
│   │   ├── api-client.ts
│   │   └── livekit.ts
│   └── components/
│       ├── game/
│       ├── chat/
│       └── ui/               # Starbucks 设计系统
├── server/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── api/
│   │   │   ├── middleware/
│   │   │   ├── admin/
│   │   │   └── ws/
│   │   ├── game/
│   │   │   ├── engine.go     # GameEngine 接口
│   │   │   ├── room.go       # 房间管理器
│   │   │   ├── packer.go     # 序列化 -> AI
│   │   │   ├── upgrade/
│   │   │   ├── doudizhu/
│   │   │   └── ...
│   │   ├── model/
│   │   ├── ai/
│   │   └── auth/
│   ├── migrations/
│   └── go.mod
└── docs/
```

## 数据库设计

### 用户与认证

```sql
CREATE TABLE users (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    nickname    VARCHAR(32) NOT NULL,
    avatar_url  VARCHAR(256),
    role        ENUM('user','admin') DEFAULT 'user',
    status      TINYINT DEFAULT 1,
    created_at  DATETIME DEFAULT NOW(),
    updated_at  DATETIME DEFAULT NOW()
);

CREATE TABLE user_auths (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id      BIGINT NOT NULL,
    provider     ENUM('google','wechat','phone','guest'),
    provider_uid VARCHAR(128) NOT NULL,
    credential   VARCHAR(256),
    UNIQUE KEY uk_provider (provider, provider_uid)
);
```

### 积分系统

```sql
CREATE TABLE scores (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    game_type   VARCHAR(32) NOT NULL,
    amount      INT NOT NULL,
    balance     INT NOT NULL,
    reason      VARCHAR(64),
    created_at  DATETIME DEFAULT NOW(),
    INDEX idx_user (user_id, created_at)
);
```

### 房间与牌局

```sql
CREATE TABLE rooms (
    id           VARCHAR(32) PRIMARY KEY,
    game_type    VARCHAR(32) NOT NULL,
    owner_id     BIGINT NOT NULL,
    status       ENUM('waiting','playing','ended') DEFAULT 'waiting',
    max_players  TINYINT DEFAULT 4,
    bot_enabled  BOOLEAN DEFAULT TRUE,
    llm_model    VARCHAR(64),
    ai_slots     TINYINT DEFAULT 0,
    created_at   DATETIME DEFAULT NOW(),
    ended_at     DATETIME
);

CREATE TABLE room_players (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id       VARCHAR(32) NOT NULL,
    user_id       BIGINT,
    is_bot        BOOLEAN DEFAULT FALSE,
    character_id  INT,
    seat_index    TINYINT NOT NULL,
    score         INT DEFAULT 0,
    status        ENUM('ready','playing','left') DEFAULT 'ready',
    UNIQUE KEY uk_room_seat (room_id, seat_index)
);

CREATE TABLE game_records (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id     VARCHAR(32) NOT NULL,
    game_type   VARCHAR(32) NOT NULL,
    round_num   INT DEFAULT 1,
    state_data  JSON,
    result      JSON,
    created_at  DATETIME DEFAULT NOW()
);
```

### 游戏操作日志（AI 分析专用，不暴露给用户）

```sql
CREATE TABLE game_actions (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    game_id      BIGINT NOT NULL,
    round_num    INT NOT NULL,
    action_seq   INT NOT NULL,
    player_id    BIGINT,
    seat_index   TINYINT NOT NULL,
    is_bot       BOOLEAN DEFAULT FALSE,
    action_type  VARCHAR(16) NOT NULL,
    cards        JSON,
    full_state   JSON,
    created_at   DATETIME(3) DEFAULT NOW(),
    INDEX idx_game (game_id, round_num, action_seq)
);
```

### 牌局快照（断线重连）

```sql
CREATE TABLE game_snapshots (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id      VARCHAR(32) NOT NULL,
    game_id      BIGINT NOT NULL,
    snapshot_at  DATETIME(3) DEFAULT NOW(),
    full_state   JSON,
    is_current   BOOLEAN DEFAULT TRUE,
    INDEX idx_game (room_id, game_id)
);
```

### LLM 配置与统计

```sql
CREATE TABLE llm_configs (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(64) NOT NULL,
    model       VARCHAR(64) NOT NULL,
    endpoint    VARCHAR(256),
    api_key     VARCHAR(256),
    priority    INT DEFAULT 0,
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  DATETIME DEFAULT NOW()
);

CREATE TABLE llm_stats (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_id   INT NOT NULL,
    game_type   VARCHAR(32) NOT NULL,
    tokens_in   INT DEFAULT 0,
    tokens_out  INT DEFAULT 0,
    latency_ms  INT DEFAULT 0,
    success     BOOLEAN DEFAULT TRUE,
    error_msg   VARCHAR(256),
    created_at  DATETIME DEFAULT NOW()
);
```

### AI 角色系统

```sql
CREATE TABLE ai_characters (
    id           INT PRIMARY KEY AUTO_INCREMENT,
    name         VARCHAR(32) NOT NULL,
    avatar_url   VARCHAR(256),
    personality  ENUM('aggressive','conservative','balanced','unpredictable') DEFAULT 'balanced',
    play_style   JSON,
    catchphrase  VARCHAR(256),
    occupation   VARCHAR(64),
    bio          TEXT,
    voice_style  ENUM('cheerful','cold','funny','gentle') DEFAULT 'cheerful',
    greeting     TEXT,
    enabled      BOOLEAN DEFAULT TRUE,
    created_at   DATETIME DEFAULT NOW()
);
```

### 数据安全分级

| 数据 | 谁可以访问 |
|------|-----------|
| 我的牌、公共牌桌信息 | 当前玩家 |
| 其他玩家的手牌 | 不允许任何玩家 API 访问 |
| `game_actions.full_state`（含所有手牌）| Go 游戏引擎内部 + AI Bot + 管理员 |
| 牌局快照 `full_state` | 断线重连时提取当前玩家手牌 + 公共信息 |

## Go 后端架构

### 分层

```
cmd/server/main.go — 配置加载、MySQL 连接池、启动 server
  └── internal/
      ├── api/          — REST + WS 路由
      │   ├── router.go
      │   ├── middleware/ (auth / admin / CORS)
      │   ├── auth.go
      │   ├── room.go
      │   ├── game.go
      │   ├── admin/   (users / rooms / llm / dashboard)
      │   └── ws/      (WebSocket Hub)
      ├── game/        — 游戏引擎
      │   ├── engine.go      (GameEngine 接口)
      │   ├── room.go        (房间管理器)
      │   ├── packer.go      (序列化 -> LLM 可读格式)
      │   ├── upgrade/
      │   ├── doudizhu/
      │   └── ...
      ├── model/       — 数据模型 + DB 操作
      ├── ai/           — LLM Bot 客户端
      └── auth/         — JWT
```

### GameEngine 接口

```go
type GameEngine interface {
    Init(players []PlayerInfo) (*GameState, error)
    ExecuteAction(action PlayerAction) (*GameState, error)
    ValidateAction(action PlayerAction) bool
    IsRoundEnd(state *GameState) bool
    CalculateScore(state *GameState) ([]PlayerScore, error)
    SerializeForAI(state *GameState) string
}
```

每种游戏实现此接口，引擎与网络层完全解耦。

### WebSocket Hub 模式

```
Client A ──WS──▶ RoomHub(goroutine) ──▶ Game Engine
Client B ──WS──▶              │              │
Client C ──WS──▶              ├──▶ DB 持久化
Client D ──Bot                 └──▶ AI 决策
```

- 每个房间一个 RoomHub goroutine，管理该房间连接
- 断线 → 标记离线 → 触发 AI 托管
- 通过 WS 广播牌局状态、聊天消息、语音状态

### AI Bot 流程

1. 轮到 AI 出牌时，引擎调用 AI 模块
2. AI 模块组装 prompt：角色设定 (`personality` + `play_style` + 口头禅等) + 序列化牌局数据
3. 调用配置的 LLM API
4. LLM 返回：出牌决策 + 聊天文本（可选）
5. 出牌通过引擎执行，聊天文本通过 WS 广播

### 断线重连流程

1. 客户端调用 `GET /api/room/:id/reconnect`
2. 服务端查询最新 `game_snapshots`
3. 从 `full_state` 提取该玩家的手牌 + 公共信息返回
4. 通过 `game_actions` 时序回放离线期间的出牌

## Next.js 前端架构

### 路由

```
app/
├── (main)/
│   ├── lobby/               # 大厅
│   ├── room/[id]/           # 游戏房间
│   ├── profile/             # 个人资料
│   └── page.tsx
├── admin/
│   ├── dashboard/
│   ├── users/
│   ├── rooms/
│   ├── llm-config/
│   ├── ai-characters/
│   └── stats/
├── auth/
│   └── login/
└── api/                     # BFF 代理
```

### 关键组件

- **牌桌 UI**：手牌、出牌区、倒计时、角色信息、聊天面板
- **聊天系统**：文本消息 + 表情包（WS 通道）、语音控件（LiveKit）
- **语音控件**：开/关麦按钮、语音活动指示器、玩家静音状态
- **设计系统**：基于 Starbucks DESIGN.md 的 UI 组件库

## 开发纪律

### TDD（测试驱动开发）

每个功能模块必须遵循 TDD 流程：

```
编写测试用例 → 运行测试（失败） → 实现代码 → 运行测试（通过） → 重构
```

- **Go 后端**：每个 GameEngine 接口实现、每个 API handler、每个 AI 模块必须有对应的单元测试
- **Go 测试范围**：游戏逻辑（发牌/出牌/判胜/算分）、房间管理、序列化、AI prompt 组装
- **Next.js 前端**：关键组件（牌桌渲染、聊天面板、语音控件）有组件测试
- **禁止**：先写实现再补测试，或跳过测试直接合并

### 严格验收标准

每个 Phase 完成前必须满足：

1. **功能验收** — 正/反例覆盖：正确场景能跑通，错误输入有合理处理，不崩溃
2. **安全验收** — 数据分级严格执行（`game_actions.full_state` 不暴露给玩家 API）
3. **无崩溃** — 异常输入、断线、并发操作不导致服务端 panic
4. **测试通过** — 该 Phase 所有测试 100% 通过；新功能引入的新测试必须通过
5. **AI Bot 验收** — Bot 出牌符合规则（不出现非法出牌），响应时间在可接受范围

每个 Phase 结束时，必须有可演示的增量成果，不积累未完成的功能到下个 Phase。

## 分阶段开发计划

### Phase 1 — 基础设施搭建

目标：可跑通的骨架，注册登录、创建房间、WS 连接
- VPS 环境准备（Go 部署、MySQL 建表、LiveKit 安装）
- Go 项目骨架（路由、中间件、MySQL 连接池、WS Hub）
- Next.js 项目初始化 + Starbucks 设计系统组件库
- 用户认证系统（登录/注册/Guest + JWT）
- Admin 基础路由框架

### Phase 2 — 第一个游戏：斗地主

目标：能完整打完一局斗地主，结算积分
- 斗地主 Game Engine（发牌、出牌、叫地主、判胜）
- WS 房间管理（创建/加入/准备/开始）
- 牌桌 UI（手牌、出牌区、倒计时）
- 游戏状态同步 + 快照持久化 + 断线重连
- 积分结算

### Phase 3 — AI Bot + 聊天/表情/语音

目标：AI 角色陪你打牌，能聊天能语音
- AI Character 角色系统（CMS 管理 + DB）
- AI 模块集成（LLM 调用 + 出牌决策 + 聊天生成）
- Bot 补位逻辑（断线/空位自动补 AI）
- 文本聊天 + 表情包（WS 通道）
- LiveKit 语音集成（开/关麦）

### Phase 4 — Admin CMS 完整功能

目标：运营管理后台完整可用
- Dashboard（在线人数、活跃房间、积分概况）
- 用户管理（列表、搜索、封禁/解封）
- 房间实时监控（查看状态、旁观/AI 替换）
- LLM 配置管理（模型切换、API Key）
- AI 角色管理（增删改、牌风/口头禅/职业等）
- LLM 调用统计（token、延迟、成功率）
- 积分手动调整

### Phase 5 — 更多游戏 + 打磨

- 打升级、炸金花、斗牛、五十K 等依次实现
- 排行榜 / 成就系统
- 邀请好友功能
- 性能优化 + 压力测试

### 依赖关系

```
Phase 1 ──▶ Phase 2 ──▶ Phase 3 ──▶ Phase 4
                │                        │
                └── Phase 5 ◀────────────┘
```
