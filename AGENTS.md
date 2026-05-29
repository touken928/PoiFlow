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
cd frontend && npx tsc --noEmit      # 前端类型检查
```

## Architecture

```
main.go + app.go         # Wails 入口 + Go/JS 桥接层
internal/
  akpool/                # AK 池（轮转、2 QPS 令牌桶限流、失败标记、WorkerCount）
  store/                 # config.yaml 读写（AK 名/Key、导出字段配置）
  task/                  # 任务队列、并发执行器、日志、CSV 续采（CompletedOps）
  exporter/              # CSV / GeoJSON 导出
pkg/
  baidu/                 # 百度 Place API v3 客户端 + BD09→WGS84
  division/              # 嵌入的行政区划数据（china.yml）
frontend/src/
  App.tsx                # 主界面（任务列表、新建、Settings、预览/表格/日志 Tab）
  PoiMap.tsx             # Leaflet 地图预览（OpenStreetMap 瓦片 + marker cluster）
  PoiTable.tsx           # POI 表格（列宽固定、溢出 Tooltip）
```

## Important gotchas

- `//go:embed all:frontend/dist` 在根包，`go test ./...` 会失败。须用 `./internal/... ./pkg/...`
- `wails build` 会覆盖 `frontend/dist/`（删除未跟踪的文件）。CI 中用 `-skipbindings`
- 修改 Go 桥接方法后须 `wails generate module` 更新 `frontend/wailsjs/`
- config.yaml 存储 AKs（命名+Key）和导出字段配置
- AK 池每 AK 限 **2 QPS**（令牌桶）。单任务并发 worker 数 = 可用 AK 数（`WorkerCount()`）
- AK 耗尽自动轮转，全部失效时**暂停**当前任务（非失败）
- 任务状态持久化到 `~/.poiflow/tasks.json`；缓存 CSV/日志在 `~/.poiflow/cache/<uuid>.{csv,log}`
- 断点续采用 `CompletedOps`（已完成查询次数）；旧任务仅有 `CompletedTargets` 时自动换算
- 并行查询靠 UID 去重，CSV 追加有 `fileMu` 保护
- 重启后运行中任务自动转为暂停
- `POIFLOW_DIR` 环境变量覆盖数据目录（默认 `~/.poiflow`）。`.env` 文件自动加载
- 前端预览依赖 Leaflet + OSM 瓦片，需联网加载地图

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
