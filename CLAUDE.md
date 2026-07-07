# drug-trials-tool 项目上下文

## 项目概述
中国临床试验数据平台，爬取 chinadrugtrials.org.cn 的试验数据，通过后台审核后在前台公开浏览。

## 架构版本
当前为 **0.2** 分支，三系统五项目架构。

## Git 分支
- `master` / `0.1`: 原始单模块代码（Go CLI + React 前端的旧版）
- `0.2`: 三系统重构（当前活跃分支）

## 目录结构

```
/ (root go.work workspace)
├── claw/           # 爬虫系统 (Go module: drug_trials_tool/claw)
│   ├── cmd/claw/   # CLI 入口
│   ├── executor/   # 爬虫执行器（可被 ctrl API 调用）
│   └── internal/
│       ├── client/ # WAF HTTP 客户端
│       ├── crawler/# 页面解析（搜索列表+详情7段）
│       └── models/ # 爬虫专用类型
├── ctrl/           # 后台系统
│   ├── api/        # Go module: drug_trials_tool/ctrl/api
│   │   └── cmd/ctrl-api/  # HTTP 服务 :8081
│   └── web/        # React + Vite :5173
│       └── src/pages/     # Login/Dashboard/CrawlManage/TrialReview/DiseaseZones/Announcements
├── spa/            # 前台系统
│   ├── api/        # Go module: drug_trials_tool/spa/api
│   │   └── cmd/spa-api/   # HTTP 服务 :8080
│   └── web/        # React + Vite :5174
│       └── src/pages/     # Home/Province/Institution/Investigator/TrialDetail/Search/DiseaseZone/Announcements
├── pkg/            # 共享模块 (Go module: drug_trials_tool/pkg)
│   ├── db/         # DB 连接 + 迁移
│   ├── models/     # 数据模型
│   └── store/      # 数据库操作
├── archive/        # 0.1 版本代码归档
├── data/           # SQLite 数据库文件
├── go.work         # Go workspace
├── ARCHITECTURE-0.2.md
└── ARCHITECTURE.md
```

## 启动方式

```bash
# 1. 后台 API (必须最先启动)
cd ctrl/api && go run .

# 2. 前台 API
cd spa/api && go run .

# 3. 后台页面
cd ctrl/web && npm install && npm run dev

# 4. 前台页面
cd spa/web && npm install && npm run dev

# 5. 爬虫 CLI（或通过 ctrl API 触发）
cd claw && go run . -keyword "胰腺" -pages 3
```

## 端口
- ctrl/api: :8081
- spa/api:  :8080
- ctrl/web: :5173
- spa/web:  :5174

## 关键数据流
1. ctrl 管理员创建病种分区 → 触发爬虫 → claw 爬取数据（published=0）
2. ctrl 管理员审核 → 公开数据（published=1）
3. spa 访问已公开数据（所有查询自带 WHERE published=1）

## 数据库核心表
- `trials`（含 published, disease_zone_id）
- `disease_zones`（病种分区）
- `announcements`（公告）
- `admin_users`（管理员）
- `provinces` / `institutions` / `investigators` / `trial_institutions`

## 技术栈
- Go 1.25, modernc.org/sqlite (纯 Go SQLite)
- React 18 + TypeScript + Vite
- ECharts 5 (中国地图)
- lucide-react (图标)

## 构建验证
所有 Go module 编译通过：
```bash
GOCACHE=".gocache" go build ./claw/...
GOCACHE=".gocache" go build ./ctrl/api/...
GOCACHE=".gocache" go build ./spa/api/...
GOCACHE=".gocache" go build ./pkg/...
```
