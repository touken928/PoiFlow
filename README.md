<h1 align="center">PoiFlow</h1>

<p align="center">
  <strong>百度地图POI批量采集工具 · 行政区划检索 · 多AK并发 · CSV导出</strong>
</p>

<p align="center">
  <a href="https://go.dev/dl/"><img src="https://img.shields.io/badge/go-1.23+-blue.svg?style=for-the-badge&logo=go" alt="Go 1.23+"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge" alt="License: MIT"></a>
  <a href="https://github.com/touken928/PoiFlow/releases"><img src="https://img.shields.io/github/v/release/touken928/PoiFlow?style=for-the-badge&logo=github" alt="GitHub release"></a>
  <a href="https://github.com/touken928/PoiFlow/stargazers"><img src="https://img.shields.io/github/stars/touken928/PoiFlow?style=for-the-badge&color=yellow&logo=github" alt="GitHub stars"></a>
</p>

> ⚠️ **合法使用声明**
> 本软件仅供学习研究使用。用户必须遵守百度地图 API 服务协议及相关法律法规，
> 不得用于任何非法用途。使用者需自行承担全部法律责任。

## 功能

- **行政区划检索**：按省/市/区县三级结构检索百度 POI 数据
- **双粒度任务**：区域粒度（选择目标级别）+ 查询粒度（实际检索细分到哪一级）
  - 例：选择省级目标，查询粒度设为区县级 → 自动展开为所有区县逐一检索
- **多搜索词**：一个任务可配置多个搜索词和分类，自动遍历所有组合
- **多 AK 管理**：支持配置多个百度 API Key（命名管理），自动轮转和限流（每 AK ≤ 3 QPS）
- **失败自动切换**：AK 额度耗尽或失效时自动切换到下一个可用 AK，全部失效时自动暂停任务
- **断点续采**：自动缓存所有结果到 `~/.poiflow/cache/`，暂停/重启后续采不重复
- **任务持久化**：任务状态持久化到磁盘，重启应用后自动恢复（运行中任务自动转为暂停）
- **坐标转换**：百度 BD09 坐标自动转换为 WGS84
- **实时日志**：日志实时推送到前端，无需手动刷新
- **Settings 面板**：Tab 式设置面板，包含 API Key 管理和关于页面

## 下载

GitHub Releases 页面提供 Windows 版本直接下载使用。

其他操作系统用户请自行编译：
- 前置要求：Go 1.23+、Node.js 18+、Wails CLI
- 克隆后执行 `wails dev` 启动开发模式，或 `wails build` 编译

## 快速使用

1. 打开软件，在左下角 **Settings** → **API Keys** → 添加百度地图 API Key（名称 + Key）
2. 点击 **新建** 创建采集任务
3. 填写任务名称，配置一个或多个搜索词和分类
4. 选择**区域粒度**（省/市/区县）并勾选具体目标
5. 选择**查询粒度**（实际API请求细分到哪一级）
6. 任务自动加入队列顺序执行，右侧面板实时显示日志
7. 支持暂停/继续/取消/删除任务
8. 任务完成后点击 **导出** 选择路径保存为 CSV 文件

## 环境变量

| 变量 | 说明 | 默认值 |
|---|---|---|
| `POIFLOW_DIR` | 数据目录（config.yaml、缓存、任务状态） | `~/.poiflow` |

可在项目根目录创建 `.env` 文件配置（自动加载）。

## 项目结构

```
PoiFlow/
├── app.go                 # Wails 桥接层
├── main.go                # 入口
├── internal/
│   ├── akpool/            # AK 池（轮转、限流、失败切换）
│   ├── akstore/           # AK 持久化（config.yaml）
│   ├── task/              # 任务系统（队列、执行、日志、CSV续采）
│   └── exporter/          # CSV 批量导出
├── pkg/
│   ├── baidu/             # 百度 Place API v3 客户端 + 坐标转换
│   └── division/          # 中国行政区划数据（内嵌 YAML）
└── frontend/
    └── src/App.tsx        # Fluent UI 前端
```

## License

MIT
