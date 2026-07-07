import React, { useEffect, useState } from "react";
import { crawlStart, crawlStatus, getCrawlLogs, getDiseaseZones, CrawlLog, DiseaseZone } from "../api";

export default function CrawlManage() {
  const [keyword, setKeyword] = useState("");
  const [pages, setPages] = useState(3);
  const [zoneId, setZoneId] = useState(0);
  const [crawling, setCrawling] = useState(false);
  const [logs, setLogs] = useState<CrawlLog[]>([]);
  const [zones, setZones] = useState<DiseaseZone[]>([]);
  const [msg, setMsg] = useState("");

  const load = () => {
    crawlStatus().then((s) => setCrawling(s.crawling));
    getCrawlLogs().then(setLogs);
    getDiseaseZones().then(setZones);
  };

  useEffect(() => { load(); const id = setInterval(load, 3000); return () => clearInterval(id); }, []);

  const handleStart = async () => {
    if (!keyword) return;
    setMsg("");
    try {
      const res = await crawlStart(keyword, pages, zoneId);
      setMsg(res.message);
      setKeyword("");
    } catch (e: any) {
      setMsg("Error: " + e.message);
    }
  };

  return (
    <div>
      <div className="page-header"><h1>爬虫管理</h1></div>

      <div className="card">
        <h3 style={{ marginBottom: "0.8rem" }}>启动爬虫</h3>
        <div className="form-row">
          <div><label>关键词</label><input placeholder="如: 胰腺" value={keyword} onChange={(e) => setKeyword(e.target.value)} /></div>
          <div><label>页数</label><input type="number" min={1} value={pages} onChange={(e) => setPages(Number(e.target.value))} /></div>
          <div><label>病种分区</label><select value={zoneId} onChange={(e) => setZoneId(Number(e.target.value))}>
            <option value={0}>不指定</option>
            {zones.map((z) => <option key={z.id} value={z.id}>{z.name}</option>)}
          </select></div>
        </div>
        <button className="btn-primary" onClick={handleStart} disabled={crawling || !keyword}>
          {crawling ? "爬取中..." : "启动爬虫"}
        </button>
        {crawling && <span className="muted" style={{ marginLeft: "0.8rem" }}><span className="spinner" style={{ display: "inline-block", width: 16, height: 16, margin: 0, verticalAlign: "middle" }} /> 爬虫正在运行...</span>}
        {msg && <p className="muted mt-1">{msg}</p>}
      </div>

      <div className="card">
        <h3 style={{ marginBottom: "0.8rem" }}>爬虫历史 ({logs.length})</h3>
        <table>
          <thead><tr><th>关键词</th><th>页数</th><th>发现</th><th>新增</th><th>状态</th><th>开始</th><th>结束</th></tr></thead>
          <tbody>
            {logs.map((l) => (
              <tr key={l.id}>
                <td>{l.keyword}</td>
                <td>{l.pages}</td>
                <td>{l.found}</td>
                <td>{l.new_items}</td>
                <td><span className={"badge " + (l.status === "running" ? "badge-yellow" : l.status === "failed" ? "" : "badge-green")}>{l.status}</span></td>
                <td className="muted">{new Date(l.start_time).toLocaleString()}</td>
                <td className="muted">{l.end_time !== l.start_time ? new Date(l.end_time).toLocaleString() : "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
