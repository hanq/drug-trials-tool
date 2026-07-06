# 系统架构方案

## 数据层次
```
省份 (Province)           ← 从详情页"各参加机构信息"的"省（州）"字段
  └─ 机构 (Institution)  ← 机构名称
       └─ 主要研究者 (PI) ← 主要研究者
            └─ 临床试验 (Trial) ← 通过 trial_investigators 关联
```

## 数据流

```
爬虫触发
  │
  ▼
搜索列表页 POST /clinicaltrials.searchlist.dhtml?keywords=<keyword>
  │ 解析 20条/页，检测总页数
  │ 去重（detail_id + data_hash）
  ▼
详情页 GET /clinicaltrials.searchlistdetail.dhtml?id=<id>
  │ 解析 collapseTwo 所有 section
  │ 提取: 7个section的结构化数据
  │       含 研究者信息(姓名/职称/Email)
  │       含 各参加机构(机构/主要研究者/省/城市)
  ▼
数据清洗 → SQLite 存储
  │
  ▼
REST API → React 前台
```

## 数据结构

```sql
-- 省
provinces (id, name, code)
  ↑ 1:n
-- 机构
institutions (id, name, province_id, city)
  ↑ 1:n
-- 主要研究者
investigators (id, name, institution_id, phone, email, title, degree)
  ↑ m:n (通过 trial_investigators 关联表)
-- 临床试验
trials (id, reg_no, title, drug_name, indication, status, applicant_name, ...)
  ↑
-- 关联表
trial_investigators (trial_id, investigator_id)
  
-- 爬虫日志
crawl_logs (id, keyword, pages, found, new, start_time, end_time, status)
```

## 文件结构

```
drug_trials_tool/
├── main.go                 # CLI入口 + API服务
├── go.mod
├── internal/
│   ├── client/             # HTTP + WAF（已有）
│   ├── crawler/            # 爬虫（扩展详情页解析）
│   ├── models/             # 数据模型
│   ├── store/              # SQLite 数据库层
│   ├── api/                # REST API 服务
│   └── cleaner/            # 数据清洗
├── frontend/               # React 前端
│   ├── package.json
│   ├── src/
│   │   ├── App.tsx
│   │   ├── pages/
│   │   │   ├── MapPage.tsx       # 中国地图首页
│   │   │   ├── ProvincePage.tsx  # 省→机构列表
│   │   │   ├── InstitutionPage.tsx # 机构→PI列表
│   │   │   ├── ExpertPage.tsx    # PI→试验列表
│   │   │   ├── TrialPage.tsx     # 试验详情
│   │   │   └── AdminPage.tsx     # 后台管理
│   │   ├── components/
│   │   │   ├── ChinaMap.tsx
│   │   │   └── SearchBar.tsx
│   │   └── types.ts
├── data/                   # SQLite 数据库
└── ANALYSIS.md
```

## API 设计

```
GET  /api/provinces                  # 所有省（含机构数量）
GET  /api/provinces/:id/institutions # 省下的机构
GET  /api/institutions/:id/investigators  # 机构下的主要研究者
GET  /api/investigators/:id/trials   # 研究者负责的试验
GET  /api/trials/:id                 # 试验详情
GET  /api/search?q=&province=&institution=&investigator=  # 搜索

GET  /api/crawl/status               # 爬取状态
POST /api/crawl/start?keyword=       # 触发爬取
```
