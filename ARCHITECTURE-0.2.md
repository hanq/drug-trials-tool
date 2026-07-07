# 0.2 架构方案 — 三系统分离

## 概览

系统拆分为三部分：

```
claw/          ctrl/                  spa/
(爬虫系统)      ├── api/ (后台接口)    ├── api/ (前台接口)
               └── web/ (后台页面)    └── web/ (前台页面)
```

## 数据流

```
    控制台指定关键词
          │
          ▼
  ┌──────────────┐     crawls      ┌──────────────────┐
  │   ctrl api   │ ──────────────→ │    claw (爬虫)    │
  │   (后台接口)  │                 │                   │
  └──────┬───────┘                 └───────┬───────────┘
         │                                 │
         │ 管理/审核                        │ 写入 unpublished 数据
         ▼                                 ▼
  ┌───────────────────────────────────────────┐
  │              SQLite 数据库                  │
  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
  │  │ trials   │  │ disease_ │  │ announce-│  │
  │  │(published│  │ zones    │  │ ments    │  │
  │  │ = false) │  │          │  │          │  │
  │  └──────────┘  └──────────┘  └──────────┘  │
  └─────────────────────────────────────────────┘
         ▲                          ▲
         │                          │
  ┌──────┴───────┐        ┌─────────┴────────┐
  │   ctrl api   │        │    spa api        │
  │  (审核/公开)  │        │  (只读已公开数据)  │
  └──────────────┘        └─────────┬────────┘
                                     │
                                     ▼
                              ┌──────────────┐
                              │  spa web     │
                              │  (前台页面)   │
                              └──────────────┘
```

## 关键设计

### 1. 数据可见性管控
- 爬虫写入的数据默认 `published = false`
- 只有 ctrl 后台审核公开后，spa 才能查看
- ctrl 可按关键词、日期、病种分区批量公开

### 2. 病种分区
- 系统初始化时由 ctrl 指定病种分区（如"胰腺癌"、"肺癌"、"乳腺癌"）
- 爬虫爬取的数据归入对应病种分区
- spa 按病种分区浏览数据

### 3. 公告系统
- ctrl 发布公告（标题、内容、发布时间、是否置顶）
- spa 展示公告列表

## 目录结构

```
drug_trials_tool/
├── claw/                        # 爬虫系统 (Go CLI)
│   ├── go.mod
│   ├── cmd/claw/main.go
│   └── internal/
│       ├── client/              # HTTP 客户端 + WAF
│       ├── crawler/             # 爬虫逻辑
│       └── models/              # 数据模型
│
├── ctrl/                        # 后台系统
│   ├── api/                     # Go 后台接口
│   │   ├── go.mod
│   │   ├── cmd/ctrl-api/main.go
│   │   └── internal/
│   │       ├── handler/
│   │       ├── store/
│   │       └── models/
│   └── web/                     # React 后台页面
│       ├── package.json
│       └── src/
│           ├── pages/           # 各功能页面
│           └── components/      # 共享组件
│
├── spa/                         # 前台系统
│   ├── api/                     # Go 前台接口
│   │   ├── go.mod
│   │   ├── cmd/spa-api/main.go
│   │   └── internal/
│   │       ├── handler/
│   │       ├── store/
│   │       └── models/
│   └── web/                     # React 前台页面
│       ├── package.json
│       └── src/
│           ├── pages/           # 各展示页面
│           └── components/      # 共享组件
│
├── data/                        # SQLite 数据库
├── docs/                        # 静态部署输出
├── ARCHITECTURE-0.2.md
└── README.md
```

## 接口设计

### ctrl api（后台接口）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST  | /api/ctrl/login                  | 管理员登录 |
| GET   | /api/ctrl/dashboard              | 后台概览统计 |
| POST  | /api/ctrl/crawl/start            | 启动爬虫 |
| GET   | /api/ctrl/crawl/status           | 爬虫状态 |
| GET   | /api/ctrl/trials                 | 试验列表（含未公开） |
| POST  | /api/ctrl/trials/:id/publish     | 公开单条试验 |
| POST  | /api/ctrl/trials/batch-publish   | 批量公开 |
| GET   | /api/ctrl/disease-zones          | 病种分区列表 |
| POST  | /api/ctrl/disease-zones          | 创建病种分区 |
| DELETE| /api/ctrl/disease-zones/:id      | 删除病种分区 |
| GET   | /api/ctrl/announcements          | 公告列表 |
| POST  | /api/ctrl/announcements          | 发布公告 |
| PUT   | /api/ctrl/announcements/:id      | 编辑公告 |
| DELETE| /api/ctrl/announcements/:id      | 删除公告 |

### spa api（前台接口）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/spa/provinces                    | 省列表（含已公开试验数） |
| GET | /api/spa/provinces/:id/institutions   | 机构列表 |
| GET | /api/spa/institutions/:id/investigators | 研究者列表 |
| GET | /api/spa/investigators/:id/trials     | 研究者试验列表 |
| GET | /api/spa/trials/:id                   | 试验详情 |
| GET | /api/spa/search                       | 搜索已公开试验 |
| GET | /api/spa/disease-zones                | 病种分区 |
| GET | /api/spa/disease-zones/:id/trials     | 分区下试验 |
| GET | /api/spa/announcements                | 公告列表 |

## 数据库设计

```sql
-- 管理员
CREATE TABLE admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 病种分区
CREATE TABLE disease_zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    keyword TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 试验（扩展 published 字段）
CREATE TABLE trials (
    detail_id TEXT PRIMARY KEY,
    reg_no TEXT, title TEXT, drug_name TEXT,
    indication TEXT, status TEXT, applicant_name TEXT,
    keyword TEXT, detail_json TEXT,
    crawl_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_hash TEXT,
    published INTEGER DEFAULT 0,
    published_at TIMESTAMP,
    published_by INTEGER REFERENCES admin_users(id),
    disease_zone_id INTEGER REFERENCES disease_zones(id)
);

-- 公告
CREATE TABLE announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_pinned INTEGER DEFAULT 0,
    published INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES admin_users(id)
);

-- 其余表复用现有设计（provinces, institutions, investigators, trial_institutions）
```

## 部署模式

```bash
# 开发模式
cd ctrl/api  && go run .     # 后台接口 :8081
cd ctrl/web  && npm run dev  # 后台页面 :5173
cd spa/api   && go run .     # 前台接口 :8080
cd spa/web   && npm run dev  # 前台页面 :5174

# 生产模式
cd ctrl && ctrl-api serve --port 8081 --static ./web/dist
cd spa  && spa-api serve  --port 8080 --static ./web/dist
```
