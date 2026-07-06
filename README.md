# 就近找临床 — Trials Nearby Info

中国临床试验机构信息可视化平台。基于《临床试验机构信息汇总表》，按省份/机构/研究者维度浏览试验分布，支持中国地图交互式查看。

## 技术栈

- **前端**: React 19 + TypeScript + Vite + ECharts
- **数据**: Python (openpyxl) 从 Excel 导出按省拆分的 JSON
- **部署**: GitHub Actions → GitHub Pages

## 快速开始

```bash
# 本地开发
cd frontend
npm install
npm run dev

# 重新导出数据
pip install openpyxl
python export_from_excel.py

# 构建
cd frontend
npm run build   # 输出到 docs/
```

## 数据结构

按省拆分，懒加载：

```
data/
├── index.json            # 35 省索引（首页用）
├── province_1..35.json   # 各省数据（机构/PI/试验）
└── overseas.json         # 25 国国际数据
```

首次加载仅 3 KB（index.json），点击省份后按需加载（10–30 KB/省）。

## 数据源

`临床试验机构信息汇总表.xlsx` — 925 行筛选后的临床试验数据，覆盖 10 个靶点、48 个试验、277 个国内外机构。

## 许可证

MIT
