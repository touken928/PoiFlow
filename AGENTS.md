# PoiFlow — AGENTS.md

## Project
Wails v2 桌面应用（Go 1.23 + React/TypeScript + Vite + Fluent UI）。
百度 POI 数据采集工具，坐标自动转换为 WGS84。

## Key commands

```bash
wails dev          # 开发模式（热重载前端）
wails build        # 生产构建（当前平台）
wails build -platform windows/amd64 -skipbindings   # 交叉编译 Windows
go test ./internal/... ./pkg/... -v  # 运行测试（排除根包的 embed）
```

## Architecture

```
main.go + app.go         # Wails 入口 + Go/JS 桥接层
internal/
  akpool/                # AK 池（轮转、3 QPS 限流、失败标记）
  store/                 # config.yaml 读写（AK 名/Key、导出字段配置）
  task/                  # 任务队列、执行器、日志、CSV 续采
  exporter/              # CSV / GeoJSON 导出
pkg/
  baidu/                 # 百度 Place API v3 客户端 + BD09→WGS84
  division/              # 嵌入的行政区划数据（china.yml）
frontend/src/App.tsx     # Fluent UI 前端（唯一前端组件）
```

## Important gotchas

- `//go:embed all:frontend/dist` 在根包，`go test ./...` 会失败。须用 `./internal/... ./pkg/...`
- `wails build` 会覆盖 `frontend/dist/`（删除未跟踪的文件）。CI 中用 `-skipbindings`
- config.yaml 存储 AKs（命名+Key）和导出字段配置
- AK 池每 AK 限 3 QPS。AK 耗尽自动轮转，全部失效时暂停当前任务
- 任务状态持久化到 `$XDG_DATA_HOME/poiflow/tasks.json`；缓存 CSV 在 `$XDG_DATA_HOME/poiflow/<uuid>.csv`
- 重启后运行中任务自动转为暂停

## Git

```
<type>: <简短描述>
```
type: `feat` `fix` `refactor` `docs` `chore` `ci` `style`

提交前检查：
1. `go test ./internal/... ./pkg/... -v`
2. `cd frontend && npx tsc --noEmit`（修改前端时）
3. `wails build`
4. **人工复核** — 所有 commit 必须人工确认无误

Push 规则：
- **只有人类可以 push**。禁止 AI 自动执行 `git push`
- Push 前确认：人工复核通过、无调试代码、无敏感文件误提交

Tag 规范：`v<major>.<minor>.<patch>`
