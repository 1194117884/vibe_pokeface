# Plan: 房间密码 + 邀请链接

**日期**: 2026-05-31
**关联 Spec**: `docs/superpowers/specs/2026-05-31-invite-link-password-room-design.md`

---

## 模块划分 & 依赖顺序

```
M1 (后端) ─────────────────────────────────────┐
   room_handler.go: 移除 is_open 过滤           │
                                                │
M2 (前端基础设施) ──────────────────────────────┤
   ws-game.ts: 新增 rejoinRoom()                │
   PasswordPrompt.tsx: 新组件                   │
                                                │
M3 (前端页面) ── 依赖 M1, M2 ──────────────────┤
   auth/login/page.tsx: redirect 支持           │
   room/doudizhu/page.tsx: 登录守卫 + 密码弹窗   │
   room/dashengji/page.tsx: 登录守卫 + 密码弹窗   │
```

## 详细任务

### M1: 后端 — room_handler.go
| 项 | 内容 |
|----|------|
| 文件 | `server/internal/api/room_handler.go` |
| 改动 | 删除 `ListRooms` 中的 `if !rm.IsOpen { continue }` 一行 |
| 验收 | `go test ./internal/api/...` 通过；大厅 API 返回全部房间 |
| 预估 | 1 行改动 |

### M2: 前端基础设施
| 项 | 内容 |
|----|------|
| M2a | `frontend/lib/ws-game.ts` — 新增 `rejoinRoom(password)` 方法（~5行） |
| M2b | `frontend/components/game/PasswordPrompt.tsx` — 新组件（~80行） |
| 验收 | TypeScript 类型检查通过；组件单测通过 |

### M3: 前端页面
| 项 | 内容 |
|----|------|
| M3a | `auth/login/page.tsx` — 支持 `?redirect=` 参数、安全校验（~15行） |
| M3b | `room/doudizhu/page.tsx` — 登录守卫 + 密码错误处理 + PasswordPrompt 集成（~40行） |
| M3c | `room/dashengji/page.tsx` — 同上（~40行） |
| 验收 | TypeScript 类型检查 + lint 通过 |

## 实施顺序

```
Step 1: M1 (后端 1行)
Step 2: M2a + M2b (前端基础设施，可并行)
Step 3: M3a (登录页)
Step 4: M3b (doudizhu 房间页)
Step 5: M3c (dashengji 房间页)
Step 6: 集成验收（全流程测试）
```

## 风险

- `router.push` 回跳时 searchParams 可能丢失 → 登录页需完整保留 `window.location.search`
- 两个房间页重复代码多 → 可抽取 `useRoomAuth` hook，但本次不做（保持最小改动）
