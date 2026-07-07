import React, { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Trial, search } from "../api";
import { FlaskConical, ChevronRight } from "lucide-react";

export default function Search() {
  const [params] = useSearchParams();
  const nav = useNavigate();
  const q = params.get("q") || "";
  const [trials, setTrials] = useState<Trial[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!q) return;
    setLoading(true);
    search(q).then((r) => { setTrials(r.trials); setTotal(r.total); setLoading(false); });
  }, [q]);

  return (
    <div>
      <div className="page-header">
        <h2>搜索: "{q}" <span className="badge">{total} 条结果</span></h2>
      </div>
      {loading && <div className="spinner" />}
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
      {!loading && trials.length === 0 && q && <p className="muted">未找到相关试验</p>}
    </div>
  );
}
