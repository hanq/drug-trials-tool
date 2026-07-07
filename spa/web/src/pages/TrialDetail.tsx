import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Trial, getTrial } from "../api";
import { ArrowLeft } from "lucide-react";

export default function TrialDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const [trial, setTrial] = useState<Trial | null>(null);
  const [err, setErr] = useState("");

  useEffect(() => {
    if (!id) return;
    getTrial(id).then(setTrial).catch((e) => setErr(e.message));
  }, [id]);

  if (err) return <div className="app-loading"><p>{err}</p></div>;
  if (!trial) return <div className="app-loading"><div className="spinner"></div></div>;

  const detail = trial.detail_json as any;
  const sections = detail?.sections || [];

  return (
    <div>
      <div className="page-header">
        <button className="btn-text" onClick={() => nav(-1)}><ArrowLeft size={16} /> 返回</button>
        <h2>试验详情</h2>
      </div>

      <div className="card">
        <table className="detail-table"><tbody>
          <tr><th>登记号</th><td>{trial.reg_no}</td></tr>
          <tr><th>试验通俗题目</th><td>{trial.title}</td></tr>
          <tr><th>药物名称</th><td>{trial.drug_name}</td></tr>
          <tr><th>适应症</th><td>{trial.indication}</td></tr>
          <tr><th>试验状态</th><td>{trial.status}</td></tr>
          <tr><th>申请人</th><td>{trial.applicant_name}</td></tr>
        </tbody></table>
      </div>

      {sections.map((sec: any, i: number) => (
        <div key={i} className="card">
          <h3 className="section-title">{sec.title}</h3>
          {sec.fields && sec.fields.map((f: any, j: number) => (
            <div key={j} className="field-row">
              <span className="field-label">{f.label}</span>
              <span className="field-value">{(f.values || []).join("; ")}</span>
            </div>
          ))}
          {sec.tables && sec.tables.map((tb: any, k: number) => (
            <div key={k} className="table-wrap">
              <table><thead><tr>
                {(tb.headers || []).map((h: string, l: number) => <th key={l}>{h}</th>)}
              </tr></thead><tbody>
                {(tb.rows || []).map((row: string[], m: number) => (
                  <tr key={m}>{(row || []).map((cell: string, n: number) => <td key={n}>{cell}</td>)}</tr>
                ))}
              </tbody></table>
            </div>
          ))}
        </div>
      ))}

      {sections.length === 0 && trial.detail_json && (
        <div className="card">
          <h3 className="section-title">完整数据</h3>
          <pre className="detail-json">{JSON.stringify(trial.detail_json, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}
