# Spec: 房间密码 + 邀请链接

**日期**: 2026-05-31
**状态**: 已确认（冻结）
**需求**: 房间支持不开放(密码)房间、邀请玩家链接、密码房自动携带密码可点击进房

---

## 1. 核心模型

| 房间类型 | `is_open` | 密码 | 大厅可见 | 加入方式 |
|---------|-----------|------|---------|---------|
| 开放房间 | `true` | 无 | ✅ | 直接加入 |
| 私有房间 | `false` | 必填 | ✅ | 大厅点击 → 房间页输密码；邀请链接(携带密码) → 直接加入 |

## 2. 用户流程

### 2.1 邀请链接流程

```
用户点击邀请链接 /room/{id}/{gameType}?password=xxx
  → 未登录 → 跳转 /auth/login?redirect=<原URL> → 登录成功 → 回跳(保留?password=)
  → 已登录 → 直接进入房间页
  → WSGameClient 自动携带 password 发起 join_room
  → 密码正确 → 正常进入房间
  → 密码错误 → 显示密码输入弹窗
```

### 2.2 大厅点击流程

```
大厅点击私有房间(hasPassword=true)
  → 跳转 /room/{id}/{gameType}（无密码参数）
  → 登录检查 → 未登录先登录
  → WSGameClient join_room 无密码
  → 服务端返回 "room password required"
  → 显示密码输入弹窗
  → 输入正确密码 → 重新 join_room 携带密码 → 进入房间
```

### 2.3 开放房间流程（不变）

```
大厅点击/邀请链接进入开放房间
  → 登录检查 → 已登录
  → WSGameClient join_room（无密码）
  → 直接进入
```

## 3. 改动清单

### 3.1 后端 (Go)

**`server/internal/api/room_handler.go`**:
- `ListRooms`: 移除 `if !rm.IsOpen { continue }` 过滤，私有房间也返回
- 保持 `HasPassword` 字段返回

### 3.2 前端 (TypeScript/React)

| 文件 | 改动 |
|------|------|
| `frontend/app/auth/login/page.tsx` | 支持 `?redirect=` 参数：登录成功后 `router.push(redirect)` 而非固定跳 `/lobby` |
| `frontend/app/(main)/room/[id]/doudizhu/page.tsx` | 1. 未登录检查 → 跳转登录页带 redirect 参数；2. 收到密码错误 → 显示 PasswordPrompt |
| `frontend/app/(main)/room/[id]/dashengji/page.tsx` | 同 doudizhu |
| `frontend/lib/ws-game.ts` | 新增 `rejoinRoom(password)` 方法（带密码重新 join_room，不改变其他状态） |
| **新** `frontend/components/game/PasswordPrompt.tsx` | 密码输入弹窗组件 |

### 3.3 PasswordPrompt 组件规格

```typescript
interface PasswordPromptProps {
  open: boolean;
  error?: string;
  loading?: boolean;
  onSubmit: (password: string) => void;
  onCancel: () => void;
}
```

- 背景遮罩不可点击关闭
- 输入为空时确认按钮禁用
- 密码错误时红色提示，不清空输入
- 取消 → 返回大厅

## 4. 验收条件

- [ ] 创建不开放(密码)房间后，该房间出现在大厅列表中，显示 🔒 图标
- [ ] 大厅点击密码房 → 进入房间页 → 弹窗提示输入密码 → 输入正确密码可进入
- [ ] 大厅点击密码房 → 输入错误密码 → 显示红色错误提示 → 可重试
- [ ] 创建不开放房间后，复制邀请链接包含 `?password=xxx`
- [ ] 邀请链接在未登录状态下打开 → 跳转登录 → 登录后自动回跳到房间页 → 自动填入密码加入
- [ ] 邀请链接在已登录状态下打开 → 直接加入房间
- [ ] 开放房间(无密码)流程不受影响
- [ ] 开放房间的邀请链接不含 password 参数
- [ ] 登录页 `redirect` 参数不为当前域名的 URL 时，忽略重定向（安全）

## 5. 不变的部分

- 房间创建逻辑不变（`is_open=false` 必填密码，`is_open=true` 清空密码）
- WebSocket 密码验证逻辑不变
- 邀请链接复制逻辑不变（已在房间页实现）
- Room 数据库模型不变
