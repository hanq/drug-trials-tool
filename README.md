# 就近找临床 — drug-trials-tool

中国临床试验数据平台。爬取 [chinadrugtrials.org.cn](https://www.chinadrugtrials.org.cn) 的试验数据，后台审核后在前台公开浏览。

## 架构 (0.2)

```
┌─────────────────────────────────────────────┐
│ claw/               ctrl/        spa/        │
│ (爬虫系统 Go CLI)    ├── api/     ├── api/   │
│                      ├── web/     └── web/   │
│                      │(后台管理)   (前台浏览) │
└─────────────────────────────────────────────┘
```

| 项目 | 路径 | 技术 | 端口 | 说明 |
|------|------|------|------|------|
| 爬虫 | `claw/` | Go CLI | - | 爬取 chinadrugtrials.org.cn |
| 后台接口 | `ctrl/api/` | Go HTTP | :8081 | 管理 API |
| 后台页面 | `ctrl/web/` | React + Vite | :5173 | 管理后台 |
| 前台接口 | `spa/api/` | Go HTTP | :8080 | 公开 API |
| 前台页面 | `spa/web/` | React + Vite | :5174 | 公开浏览 |

## 数据流

1. 管理员在 ctrl 创建**病种分区**（如"胰腺癌"）
2. 通过 ctrl 触发**爬虫** → claw 爬取数据入库（默认不可见，published=0）
3. ctrl **审核**后公开 → published=1
4. **病种分区切换** — spa 前台 Header 下拉框选择病种，数据按分区过滤；
   默认"全部病种"，选区后省份/机构/研究者/搜索均限定范围
5. spa 浏览器**只读**已公开数据

## 数据库

- `data/db/trials.db` — SQLite，已完成 0.2 架构迁移
- 含 admin_users / disease_zones / announcements / crawl_logs 等新表
- 已有 197 条试验归入"胰腺癌"分区（disease_zone_id=1），published=0（默认不公开）
- 默认管理员：admin / admin123（SHA256 验证）

## 构建验证

```bash
# Go 后端
GOCACHE=".gocache" go build ./pkg/... ./spa/api/... ./ctrl/api/... ./claw/...
go vet ./pkg/... && go vet ./spa/api/...

# 前端 spa
cd spa/web && npm install && npm run build

# 前端 ctrl
cd ctrl/web && npm install && npm run build
```

## Git 分支

- `master` / `0.1`: 原始单模块代码
- `0.2`: 三系统重构（当前）

## 快速开始

```bash
# 后台 API
cd ctrl/api && go run .

# 前台 API
cd spa/api && go run .

# 后台页面
cd ctrl/web && npm install && npm run dev

# 前台页面
cd spa/web && npm install && npm run dev

# 爬虫 CLI
cd claw && go run . -keyword "胰腺" -pages 3
```

## 许可证

MIT
