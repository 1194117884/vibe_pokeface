# P0-A: Fix ValidateAction + Remove Dead PassCount Field

> Part of 斗地主规则对齐 — 出牌校验和代码清理

## Scope

2 files, bug fix + cleanup.

## Changes

### 1. engine.go — ValidateAction 补 CanBeat 检查

**Problem:** `ValidateAction` 在 playing 阶段校验 `play` 动作时，只检查了牌型有效性和手牌归属，没有检查能否压过 `LastPlay`。调用方（前端/测试）通过 `ValidateAction` 预校验会得到假的 `true`，但实际执行时 `handlePlay` 会因 `CanBeat` 失败而报错。

**Fix:** 在 `ValidateAction` 的 play 分支末尾、`return true` 之前，增加：

```go
if gs.LastPlay != nil && !CanBeat(play, gs.LastPlay.Play) {
    return false
}
```

首出（LastPlay==nil）不需要压牌检查，保持原逻辑。

### 2. state.go — 删除死字段 PassCount

**Problem:** `GameState` 有两个 pass 计数字段：`PassCount` 和 `ConsecutivePasses`。实际逻辑仅使用 `ConsecutivePasses`，`PassCount` 从未被读写。

**Fix:** 删除 `PassCount int \`json:"pass_count"\`` 行。

## Acceptance

```bash
cd server && go vet ./... && go test ./internal/game/doudizhu/ -v
```

所有已有测试通过。无 vet 错误。

## Dependencies

P0-B 的前置步骤（P0-B 是重构叫地主/抢地主两阶段）。
