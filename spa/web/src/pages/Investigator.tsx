import React, { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Trial, getTrials } from "../api";
import { FlaskConical, ArrowLeft, ChevronRight } from "lucide-react";

export default function InvestigatorPage() {
  const [params] = useSearchParams();
  const nav = useNavigate();
  const name = params.get("name") || "";
  const iid = parseInt(params.get("iid") || "0");
  const [trials, setTrials] = useState<Trial[]>([]);

  useEffect(() => { if (name && iid) getTrials(name, iid).then(setTrials); }, [name, iid]);

  return (
    <div>
      <div className="page-header">
        <button className="btn-text" onClick={() => nav(-1)}><ArrowLeft size={16} /> 返回</button>
        <h2>{name} <span className="badge">{trials.length} 项试验</span></h2>
      </div>
      {trials.map((t) => (
        <div key={t.detail_id} className="card card-click" onClick={() => nav("/trial/" + t.detail_id)}>
          <div className="card-body">
            <FlaskConical size={18} className="icon" />
            <div className="trial-info">
              <strong>{t.title}</strong>
              <span className="muted">{t.indication?.slice(0, 60)}</span>
            </div>
            <ChevronRight size={16} className="chevron" />
          </div>
        </div>
      ))}
      {trials.length === 0 && <p className="muted">暂无试验数据</p>}
    </div>
  );
}
