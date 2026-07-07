import React, { useEffect, useState } from "react";
import { getDashboard, DashboardStats, CrawlLog } from "../api";

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [err, setErr] = useState("");

  useEffect(() => {
    getDashboard().then(setStats).catch((e) => setErr(e.message));
  }, []);

  if (err) return <div className="page-header"><h1>Dashboard</h1><p style={{ color: "red" }}>{err}</p></div>;
  if (!stats) return <div className="spinner" />;

  const cards = [
    { label: "试验总数", value: stats.total_trials },
    { label: "已公开", value: stats.published_trials, cls: "badge-green" },
    { label: "待审核", value: stats.pending_trials, cls: "badge-yellow" },
    { label: "省份", value: stats.total_provinces },
    { label: "机构", value: stats.total_institutions },
    { label: "研究者", value: stats.total_investigators },
    { label: "病种分区", value: stats.total_zones },
  ];

  return (
    <div>
      <div className="page-header"><h1>Dashboard</h1></div>
      <div className="stats-grid">
        {cards.map((c) => (
          <div key={c.label} className="card">
            <h3>{c.label}</h3>
            <div className="value">{c.value}</div>
          </div>
        ))}
      </div>
      {stats.recent_crawls.length > 0 && (
        <div className="card">
          <h3 style={{ marginBottom: "0.8rem" }}>最近爬虫记录</h3>
          <table>
            <thead><tr><th>关键词</th><th>页数</th><th>发现</th><th>新增</th><th>状态</th><th>时间</th></tr></thead>
            <tbody>
              {stats.recent_crawls.map((l) => (
                <tr key={l.id}>
                  <td>{l.keyword}</td>
                  <td>{l.pages}</td>
                  <td>{l.found}</td>
                  <td>{l.new_items}</td>
                  <td><span className={"badge " + (l.status === "running" ? "badge-yellow" : l.status === "failed" ? "" : "badge-green")}>{l.status}</span></td>
                  <td className="muted">{new Date(l.start_time).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
