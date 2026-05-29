<p align="center">
  <img src="build/appicon.png" width="80" height="80" alt="logo"><br>
  <strong>PoiFlow</strong><br>
  <em>百度POI数据采集工具 · 行政区划检索 · 多AK并发 · 地图预览 · CSV/GeoJSON导出</em>
</p>

<p align="center">
  <a href="https://go.dev/dl/"><img src="https://img.shields.io/badge/go-1.23+-blue?style=flat&logo=go" alt="Go"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat" alt="MIT"></a>
  <a href="https://github.com/touken928/PoiFlow/releases"><img src="https://img.shields.io/github/v/release/touken928/PoiFlow?style=flat&logo=github" alt="release"></a>
  <a href="https://github.com/touken928/PoiFlow/stargazers"><img src="https://img.shields.io/github/stars/touken928/PoiFlow?style=flat&color=yellow&logo=github" alt="stars"></a>
</p>

> 本软件仅供学习研究。使用者须遵守百度地图API服务协议及相关法律法规。

## 功能

- **行政区划检索** — 省/市/区县三级结构检索百度 POI，坐标自动转换 WGS84
- **多搜索词** — 一个任务可配置多个搜索词和分类，支持从文件批量导入
- **多 AK 并发** — 命名管理多个 API Key；单任务按可用 AK 数并行查询（每 Key 约 2 QPS），失败自动切换
- **断点续采** — 结果实时写入磁盘，按查询进度续采，UID 去重不重复
- **任务持久化** — 状态持久化到磁盘，重启自动恢复，运行中任务转为暂停
- **双精度查询** — 目标可细化到省/市/区县任意层级执行检索
- **预览 / 表格 / 日志** — 选中任务后默认 Leaflet 地图预览 POI，可切换表格浏览与实时日志
- **CSV / GeoJSON 导出** — 支持自定义导出字段

## 使用

Windows 用户从 [Releases](https://github.com/touken928/PoiFlow/releases) 下载。其他系统需 Go 1.23+、Node.js 18+、Wails CLI：

```bash
git clone https://github.com/touken928/PoiFlow.git
cd PoiFlow
wails dev      # 开发模式
wails build    # 构建当前平台
```

![主界面](./assets/面板.png)

1. 左下角 **Settings** → **API Keys** 添加百度 API Key（配置多个 Key 可提升单任务并发）
2. 点击 **新建**，填写搜索词和分类（至少一个，支持从“搜索词，分类”的文本文件导入），从树形列表选择目标区域
3. 任务自动排队执行；选中任务后右侧面板默认 **预览** 地图，可切换 **表格** / **日志**；窗口标题显示当前任务名
4. 暂停/继续/重试/删除任务，完成后导出 CSV 或 GeoJSON

## 数据目录

默认 `~/.poiflow/`，可通过环境变量 `POIFLOW_DIR` 覆盖：

| 路径 | 说明 |
|------|------|
| `config.yaml` | AK 与导出字段配置 |
| `tasks.json` | 任务状态 |
| `cache/<uuid>.csv` | 任务采集缓存 |
| `cache/<uuid>.log` | 任务日志 |

## 项目结构

```
app.go + main.go          # Wails 桥接 + 入口
internal/akpool/           # AK 池（轮转、令牌桶限流、失败切换、并发 worker 数）
internal/store/            # config.yaml 读写（AK 配置 + 导出字段）
internal/task/             # 任务队列、并发执行器、日志、续采
internal/exporter/         # CSV / GeoJSON 导出
pkg/baidu/                 # 百度 Place API v3 客户端 + BD09→WGS84
pkg/division/              # 内嵌行政区划数据（china.yml）
frontend/src/App.tsx         # 主界面
frontend/src/PoiMap.tsx      # Leaflet 地图预览
frontend/src/PoiTable.tsx    # POI 表格
```

MIT License
