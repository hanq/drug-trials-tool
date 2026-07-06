---
name: trials-nearby-info
description: 中国临床试验信息可视化项目（就近找临床）。基于药物临床试验登记平台数据，按省份/机构/研究者维度展示试验分布。包含数据采集（爬虫+下载）、清洗、导出、前端可视化的完整管道。适用场景：查本项目架构、数据导出流程、前端构建、GitHub Pages部署、已知问题和修复。
---

# Trials Nearby Info (就近找临床)

## 项目结构

* `main.go` - CLI 入口，支持三种模式：
  1. `-crawl <keyword> -crawl-pages N` - 从 chinadrugtrials.org.cn 爬取试验列表
  2. `-keyword <k> -output <dir>` - 下载试验详情
  3. `-export <path>` - 清洗数据并导出为 JSON（供前端使用）
* `internal/store/` - SQLite 存储层
* `internal/crawler/` - 爬虫实现
* `internal/downloader/` - 试验详情下载器
* `internal/client/` - HTTP 客户端
* `frontend/` - React + Vite + TypeScript 前端
* `data/db/trials.db` - SQLite 数据文件

## 数据管道

1. 爬取: `-crawl 胰腺 -crawl-pages 10`
2. 下载: `-keyword 胰腺 -output ./pancreas_docs`
3. 导出: `-export frontend/public/data.json`
4. 构建: `frontend/` 目录下 `npx vite build`

## 关键 Bug 修复 (2026-07-05)

**问题**: 导出函数构建了 `instList` 和 `invList`，但未赋值到 `exportData` map。
**修复**: 在 `main.go` 中 `exportData["trials"] = trialList` 后增加两行：
```
exportData["institutions"] = instList
exportData["investigators"] = invList
```

## 前端部署 (GitHub Pages)

* 构建产物在 `frontend/dist/`
* 独立推送到 `git@github.com:hanq/trials-nearby-info.git`
* 当前沙箱环境 SSH 端口 22/443 被阻止，推送到 GitHub 需使用 HTTPS Token 或手动推送

## 已知数据状态

* 当前数据库: 173 项试验（胰腺相关），31 省份，439 机构，1079 研究者
* `frontend/public/data.json` - 前端使用的 JSON（已包含完整 5 个数组：provinces, institutions, investigators, trials, trial_institutions）

## 构建命令

```
go build -buildvcs=false -o drug_trials_tool.exe .
.\drug_trials_tool.exe -export frontend/public/data.json
cd frontend; npx vite build
```
