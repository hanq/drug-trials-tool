import React, { useEffect, useState } from "react";
import { getTrials, publishTrial, batchPublish, getDiseaseZones, Trial, DiseaseZone } from "../api";

export default function TrialReview() {
  const [trials, setTrials] = useState<Trial[]>([]);
  const [zones, setZones] = useState<DiseaseZone[]>([]);
  const [filterZone, setFilterZone] = useState(0);
  const [filterKw, setFilterKw] = useState("");
  const [filterPub, setFilterPub] = useState(0);
  const [selected, setSelected] = useState<Trial | null>(null);
  const [msg, setMsg] = useState("");

  const load = () => {
    getTrials(filterZone || undefined, filterKw || undefined, filterPub === -1 ? undefined : filterPub).then(setTrials);
    getDiseaseZones().then(setZones);
  };

  useEffect(() => { load(); }, [filterZone, filterKw, filterPub]);

  const handlePublish = async (id: string) => {
    try { await publishTrial(id); setMsg("已公开"); load(); } catch (e: any) { setMsg(e.message); }
  };

  const handleBatch = async () => {
    try {
      const r = await batchPublish(filterZone || undefined, filterKw || undefined);
      setMsg("批量公开: " + JSON.stringify(r));
      load();
    } catch (e: any) { setMsg(e.message); }
  };

  const statusLabel = filterPub === -1 ? "全部" : filterPub === 1 ? "已公开" : "隐藏";
  const hiddenCount = trials.filter(t => t.published === 0).length;
  const showCount = filterPub === -1 ? `${trials.length} 条` : `${trials.length} 条${filterPub === 0 ? `（${hiddenCount} 条待公开）` : ""}`;

  return (
    <div>
      <div className="page-header"><h1>试验管理 <span className="badge badge-yellow">{showCount}</span></h1></div>

      <div className="card">
        <div className="form-row">
          <div><label>病种分区</label><select value={filterZone} onChange={(e) => setFilterZone(Number(e.target.value))}>
            <option value={0}>全部</option>
            {zones.map((z) => <option key={z.id} value={z.id}>{z.name}</option>)}
          </select></div>
          <div><label>关键词</label><input placeholder="筛选关键词" value={filterKw} onChange={(e) => setFilterKw(e.target.value)} /></div>
          <div><label>状态</label><select value={filterPub} onChange={(e) => setFilterPub(Number(e.target.value))}>
            <option value={0}>隐藏</option>
            <option value={1}>公开</option>
            <option value={-1}>全部</option>
          </select></div>
          <div style={{ display: "flex", alignItems: "flex-end" }}>
            <button className="btn-success" onClick={handleBatch}>批量公开当前筛选</button>
          </div>
        </div>
        {msg && <p className="muted">{msg}</p>}
      </div>

      <table>
        <thead><tr><th>登记号</th><th>标题</th><th>药物</th><th>适应症</th><th>关键词</th><th>状态</th><th>爬取时间</th><th>操作</th></tr></thead>
        <tbody>
          {trials.map((t) => (
            <tr key={t.detail_id} onClick={() => setSelected(t)} style={{ cursor: "pointer" }}>
              <td>{t.reg_no}</td>
              <td style={{ maxWidth: 250, overflow: "hidden", textOverflow: "ellipsis" }}>{t.title}</td>
              <td>{t.drug_name}</td>
              <td style={{ maxWidth: 180, overflow: "hidden", textOverflow: "ellipsis" }}>{t.indication}</td>
              <td><span className="badge badge-gray">{t.keyword}</span></td>
              <td><span className={"badge " + (t.published ? "badge-green" : "badge-yellow")}>{t.published ? "公开" : "隐藏"}</span></td>
              <td className="muted">{new Date(t.crawl_time).toLocaleDateString()}</td>
              <td>{t.published ? <span className="muted">-</span> : <button className="btn-sm btn-success" onClick={(e) => { e.stopPropagation(); handlePublish(t.detail_id); }}>公开</button>}</td>
            </tr>
          ))}
          {trials.length === 0 && <tr><td colSpan={8} className="muted" style={{ textAlign: "center", padding: "2rem" }}>没有符合条件的试验</td></tr>}
        </tbody>
      </table>

      {selected && (
        <div className="modal-overlay" onClick={() => setSelected(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>{selected.reg_no}</h2>
            <p className="muted mb-1" style={{ color: selected.published ? "var(--success)" : "var(--warn)" }}>{selected.published ? "已公开" : "隐藏"}</p>
            <table className="mb-1"><tbody>
              <tr><th>登记号</th><td>{selected.reg_no}</td></tr>
              <tr><th>药物名称</th><td>{selected.drug_name}</td></tr>
              <tr><th>适应症</th><td>{selected.indication}</td></tr>
              <tr><th>状态</th><td>{selected.status}</td></tr>
              <tr><th>申请人</th><td>{selected.applicant_name}</td></tr>
              <tr><th>关键词</th><td>{selected.keyword}</td></tr>
            </tbody></table>
            <div className="actions">
              {!selected.published && <button className="btn-success btn-sm" onClick={() => { handlePublish(selected.detail_id); setSelected(null); }}>公开</button>}
              <button className="btn-ghost btn-sm" onClick={() => setSelected(null)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
