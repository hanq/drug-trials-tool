import React, { useEffect, useState } from "react";
import { getTrials, publishTrial, batchPublish, getDiseaseZones, Trial, DiseaseZone } from "../api";

export default function TrialReview() {
  const [trials, setTrials] = useState<Trial[]>([]);
  const [zones, setZones] = useState<DiseaseZone[]>([]);
  const [filterZone, setFilterZone] = useState(0);
  const [filterKw, setFilterKw] = useState("");
  const [selected, setSelected] = useState<Trial | null>(null);
  const [msg, setMsg] = useState("");

  const load = () => {
    getTrials(filterZone || undefined, filterKw || undefined).then(setTrials);
    getDiseaseZones().then(setZones);
  };

  useEffect(() => { load(); }, [filterZone, filterKw]);

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

  return (
    <div>
      <div className="page-header"><h1>试验审核 <span className="badge badge-yellow">{trials.length} 条待审核</span></h1></div>

      <div className="card">
        <div className="form-row">
          <div><label>病种分区</label><select value={filterZone} onChange={(e) => setFilterZone(Number(e.target.value))}>
            <option value={0}>全部</option>
            {zones.map((z) => <option key={z.id} value={z.id}>{z.name}</option>)}
          </select></div>
          <div><label>关键词</label><input placeholder="筛选关键词" value={filterKw} onChange={(e) => setFilterKw(e.target.value)} /></div>
          <div style={{ display: "flex", alignItems: "flex-end" }}>
            <button className="btn-success" onClick={handleBatch}>批量公开当前筛选</button>
          </div>
        </div>
        {msg && <p className="muted">{msg}</p>}
      </div>

      <table>
        <thead><tr><th>登记号</th><th>标题</th><th>药物</th><th>适应症</th><th>关键词</th><th>爬取时间</th><th>操作</th></tr></thead>
        <tbody>
          {trials.map((t) => (
            <tr key={t.detail_id} onClick={() => setSelected(t)} style={{ cursor: "pointer" }}>
              <td>{t.reg_no}</td>
              <td style={{ maxWidth: 300, overflow: "hidden", textOverflow: "ellipsis" }}>{t.title}</td>
              <td>{t.drug_name}</td>
              <td style={{ maxWidth: 200, overflow: "hidden", textOverflow: "ellipsis" }}>{t.indication}</td>
              <td><span className="badge badge-gray">{t.keyword}</span></td>
              <td className="muted">{new Date(t.crawl_time).toLocaleDateString()}</td>
              <td><button className="btn-sm btn-success" onClick={(e) => { e.stopPropagation(); handlePublish(t.detail_id); }}>公开</button></td>
            </tr>
          ))}
          {trials.length === 0 && <tr><td colSpan={7} className="muted" style={{ textAlign: "center", padding: "2rem" }}>没有待审核的试验</td></tr>}
        </tbody>
      </table>

      {selected && (
        <div className="modal-overlay" onClick={() => setSelected(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>{selected.title}</h2>
            <table className="mb-1"><tbody>
              <tr><th>登记号</th><td>{selected.reg_no}</td></tr>
              <tr><th>药物名称</th><td>{selected.drug_name}</td></tr>
              <tr><th>适应症</th><td>{selected.indication}</td></tr>
              <tr><th>状态</th><td>{selected.status}</td></tr>
              <tr><th>申请人</th><td>{selected.applicant_name}</td></tr>
              <tr><th>关键词</th><td>{selected.keyword}</td></tr>
            </tbody></table>
            <div className="actions">
              <button className="btn-success btn-sm" onClick={() => { handlePublish(selected.detail_id); setSelected(null); }}>公开</button>
              <button className="btn-ghost btn-sm" onClick={() => setSelected(null)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
