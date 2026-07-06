# chinadrugtrials.org.cn 数据抓取分析文档

## 1. 整体架构

```
搜索列表页                   详情页                      下载
───────────────────────────────────────────────────────────────
GET/POST                     GET                          POST
/clinicaltrials.             /clinicaltrials.             /clinicaltrials.
searchlist.dhtml             searchlistdetail.dhtml       searchlistdetail.dhtml
                                                          ?_export=doc
```

## 2. 反爬机制 (WAF)

网站使用 FSSBBI 系 WAF（JS Challenge 型）：

| Cookie | 说明 |
|--------|------|
| `FSSBBIl1UgzbN7N80S` | 会话标识，首次请求返回 |
| `FSSBBIl1UgzbN7N80T` | 挑战签名，JS 执行后生成 |
| `token` | 登录令牌，非必需 |

**绕过方法**（参考 `drug_tool` 项目的模式）：
1. 首次 GET 请求页面 → 服务器返回 202 + 混淆 JS challenge
2. 忽略 JS 内容，仅提取 set-cookie 中的 `FSSBBIl1UgzbN7N80S`
3. sleep 2-3 秒
4. 后续请求自动携带该 cookie，WAF 放行
5. 如果再次遇到 202，重新 init session 再重试

注意：JS challenge 不需要实际执行，WAF 在第一次返回 202 时就已经发出了有效 cookie，sleep 等待后即可正常使用。

## 3. 搜索列表页

### 3.1 请求方式

搜索表单结构：

```html
<form id="searchfrm" method="post" action="/clinicaltrials.searchlist.dhtml">
  <input type="hidden" id="currentpage" name="currentpage" value="1"/>
  <input type="hidden" id="sort" name="sort" value="desc"/>
  <input type="hidden" id="sort2" name="sort2" value=""/>
  <input type="hidden" id="rule" name="rule" value="CTR"/>
  <input type="hidden" id="secondLevel" name="secondLevel" value="0"/>
  <input type="hidden" id="id" name="id" value=""/>
  <input type="hidden" id="ckm_index" name="ckm_index" value=""/>
  <input type="text" name="keywords" value="胰腺"/>
</form>
```

**两种搜索方式**：

| 方式 | URL | 说明 |
|------|-----|------|
| GET（快速搜索） | `?keywords=胰腺` | 首页搜索框回车触发 |
| POST（分页搜索） | 提交完整表单 | 点击"查询"按钮、翻页时触发 |

POST 是实现分页的关键，需要提交所有隐藏字段。

### 3.2 搜索结果 HTML 结构

```html
<table>
  <tr class="Tab_title">
    <th width="7%">序号</th>
    <th width="17%">登记号</th>
    <th width="17%">试验状态</th>
    <th width="18%">药物名称</th>
    <th width="21%">适应症</th>
    <th width="20%">试验通俗题目</th>
  </tr>
  <tr style=" color:#535353">
    <td>&nbsp;1</td>
    <td><a href="javascript:void(0)" onclick="getDetail(this.id)" id="32hexID" name="1">CTR20262592</a></td>
    <td><a ... id="32hexID" name="1">进行中 招募中</a></td>
    <td><a ... id="32hexID" name="1">小儿银连颗粒</a></td>
    <td><a ... id="32hexID" name="1">适应症名称</a></td>
    <td><a ... id="32hexID" name="1">试验通俗题目</a></td>
  </tr>
</table>
```

关键点：
- 每行所有 `<a>` 标签共享同一个 `id`（32 位 hex），这是该试验的唯一标识
- `name` 属性表示行号
- 列顺序固定：序号 → 登记号 → 试验状态 → 药物名称 → 适应症 → 试验通俗题目

### 3.3 分页

```html
当前第 <i>1</i> 页，共 <i>10</i> 页，共 <i>197</i> 条记录

<ul class="pagination">
  <li class="active"><a href="#" onclick="gotopage(1)">1</a></li>
  <li><a href="#" onclick="gotopage(2)">2</a></li>
  ...
  <li><a href="#" onclick="gotopage(10)">10</a></li>
</ul>
```

翻页通过 `gotopage(N)` JS 函数实现，最终提交 POST 到 `/clinicaltrials.searchlist.dhtml`，参数 `currentpage=N`。

**分页数据的获取**：
- 发送 POST 请求，参数 `currentpage=1,2,3...` 及其他固定字段
- 每页约 20 条记录
- 从页面中的 `共 X 页` 解析总页数

## 4. 详情页

### 4.1 访问方式

```http
GET /clinicaltrials.searchlistdetail.dhtml?id=6728595b4eea44dfa816c8cdfb42c147
```

### 4.2 HTML 结构（折叠面板）

详情页使用 Bootstrap 的 collapse 组件，包含两个折叠面板：

```
详情页
├── collapseOne (默认展开 collapse in)
│   └── 基本信息
│       ├── 登记号
│       ├── 试验状态
│       ├── 药物名称
│       ├── 适应症
│       ├── 试验通俗题目
│       └── ...（精简信息）
│
├── collapseTwo (默认折叠 collapse，点击"+"展开)
│   └── 公示的试验信息
│       ├── 一、题目和背景信息
│       │   ├── 登记号
│       │   ├── 相关登记号
│       │   ├── 药物名称
│       │   ├── 药物类型
│       │   ├── 临床申请受理号
│       │   ├── 适应症
│       │   ├── 试验专业题目
│       │   ├── 试验通俗题目
│       │   ├── 试验方案编号
│       │   ├── 方案最新版本号
│       │   ├── 版本日期
│       │   └── 方案是否为联合用药
│       │
│       ├── 二、申请人信息
│       │   ├── 申请人名称（可多个）
│       │   └── ...
│       │
│       ├── 三、临床试验信息
│       │   ├── 试验目的
│       │   ├── 试验分类
│       │   ├── 试验分期
│       │   ├── 设计类型
│       │   ├── 随机化
│       │   ├── 盲法
│       │   ├── 试验范围
│       │   ├── 受试者年龄
│       │   ├── 受试者性别
│       │   ├── 是否健康受试者
│       │   ├── 入选标准
│       │   ├── 排除标准
│       │   ├── 主要终点指标
│       │   ├── 次要终点指标
│       │   ├── 试验药
│       │   ├── 对照药
│       │   └── ...
│       │
│       ├── 四、研究者信息
│       │   ├── 主要研究者信息
│       │   └── ...
│       │
│       ├── 五、伦理委员会信息
│       │   ├── 伦理委员会名称
│       │   ├── 审查结论
│       │   └── ...
│       │
│       └── 六、各研究机构信息
│           ├── 机构名称（可多个）
│           ├── 主要研究者
│           └── ...
│
└── 下载按钮
    └── POST /clinicaltrials.searchlistdetail.dhtml?_export=doc
        表单数据: id=32hexID
        返回: .doc 文件
```

### 4.3 关键发现：数据已在 HTML 中，无需 JS 执行

`collapseTwo` 的内容是**服务端渲染**的，直接包含在 HTML 中。Bootstrap 的 collapse 组件只是通过 CSS `display: none/block` 切换可见性：

```html
<!-- 默认隐藏 -->
<div id="collapseTwo" class="panel-collapse collapse" role="tabpanel">
  <div class="panel-body">
    <!-- 所有试验数据都在这里 -->
    <div class="searchDetailPartTit">一、题目和背景信息</div>
    <table class="searchDetailTable">
      <tr><th>登记号</th><td>CTR20262230</td></tr>
      ...
    </table>
  </div>
</div>
```

浏览器中点击"+"按钮只是触发：
```javascript
// 切换 CSS class，无 AJAX 请求
$("#collapseTwo").collapse("toggle");
```

因此，**提取这些数据不需要 JS 执行，也不需要 headless 浏览器**。直接解析 HTML 即可。

### 4.4 数据提取方法

每个 section 的结构模式：

```html
<div class="searchDetailPartTit">一、题目和背景信息</div>
<table class="searchDetailTable">
  <tr>
    <th>登记号</th>
    <td colspan="3">CTR20262230</td>
  </tr>
  <tr>
    <th width="15%">试验方案编号</th>
    <td width="35%">DN022150-302</td>
    <th width="15%">方案最新版本号</th>
    <td width="35%">V1.0</td>
  </tr>
</table>
```

支持两种表格布局：
- **单列模式**：`<th>` + `<td colspan="3">` — 一个 label 对应一个值
- **双列模式**：`<th><td><th><td>` — 两个 label-value 对同行排列

## 5. 下载文档

```http
POST /clinicaltrials.searchlistdetail.dhtml?_export=doc
Content-Type: application/x-www-form-urlencoded

id=6728595b4eea44dfa816c8cdfb42c147
```

返回文档类型由 magic bytes 自动检测：
| Magic Bytes | 类型 |
|-------------|------|
| `D0CF11E0` | .doc (OLE2) |
| `PK\x03\x04` | .docx (Office Open XML) |
| `%PDF` | .pdf |
| `<!DOCT` | .html (WAF 回退) |

## 6. 参数总结

| 命令行参数 | 用途 | 默认值 |
|-----------|------|--------|
| `-keyword` | 搜索关键词（如"胰腺"） | 空 |
| `-output` | 文档输出目录 | `./downloads` |
| `-max` | 最大下载数（0=全部） | 10 |
| `-pages` | 搜索结果页数 | 1（约20条/页） |
| `-ids` | 直接通过ID下载 | 空 |
| `-cookies` | 注入预设cookie | 空 |
| `-save-html` | 保存HTML用于调试 | 空 |
