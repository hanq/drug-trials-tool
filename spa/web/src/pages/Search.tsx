import React, { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Trial, search } from "../api";
import { useZone } from "../ZoneContext";
import { FlaskConical, ChevronRight } from "lucide-react";

export default function Search() {
  const [params] = useSearchParams();
  const nav = useNavigate();
  const q = params.get("q") || "";
  const zoneIdParam = params.get("zone_id");
  const zoneId = zoneIdParam ? parseInt(zoneIdParam, 10) : undefined;
  const { zones } = useZone();
  const zoneName = zoneId ? zones.find((z) => z.id === zoneId)?.name : "";
  const [trials, setTrials] = useState<Trial[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!q) return;
    setLoading(true);
    search(q, 1, zoneId).then((r) => { setTrials(r.trials); setTotal(r.total); setLoading(false); });
  }, [q, zoneId]);

  const titleSuffix = zoneName ? ` (${zoneName})` : "";

  return (
    <div>
      <div className="page-header">
        <h2>搜索: "{q}"{titleSuffix} <span className="badge">{total} 条结果</span></h2>
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
