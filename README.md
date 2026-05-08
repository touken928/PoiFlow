# PoiFlow

百度 POI 数据采集桌面工具，基于 [Wails](https://wails.io) + [Fluent UI](https://react.fluentui.dev) 构建。

## 功能

- **行政区划检索**：按省/市/区县三级结构检索百度 POI 数据
- **双粒度任务**：区域粒度（选择目标级别） + 查询粒度（实际检索细分到哪一级）
  - 例：选择省级目标，查询粒度设为区县级 → 自动展开为所有区县逐一检索
- **多 AK 管理**：支持同时配置多个百度 API Key，自动轮转和限流（每 AK ≤ 3 QPS）
- **失败自动切换**：AK 额度耗尽或失效时自动切换到下一个可用 AK
- **坐标转换**：百度 BD09 坐标自动转换为 WGS84
- **断点续采**：指定 CSV 导出路径后实时追加写入，重复运行自动跳过已有 UID
- **日志监控**：实时查看每条检索任务的执行日志

## 快速开始

### 前置要求

- Go 1.23+
- Node.js 18+
- Wails CLI

### 安装

```bash
git clone https://github.com/touken928/PoiFlow.git
cd PoiFlow
wails dev
```

### 配置 API Key

在软件左下角 **Settings** → 添加百度地图 API Key。首次使用也可通过环境变量设置：

```bash
export BAIDU_AK=your_ak_here   # 多个用逗号分隔
```

### 创建任务

1. 点击 **新建** 打开任务对话框
2. 填写搜索词（如：ATM机）和可选分类（如：银行）
3. 选择**区域粒度**（省/市/区县）并勾选目标
4. 选择**查询粒度**（实际API请求细分到哪一级）
5. 可选：指定 CSV 路径以支持实时写入和断点续采
6. 点击创建，任务自动加入队列顺序执行

### 构建

```bash
# 当前平台
wails build

# 交叉编译 Windows 版本
wails build -platform windows/amd64
```

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
