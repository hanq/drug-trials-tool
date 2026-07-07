import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Trial, getZoneTrials, getDiseaseZones, DiseaseZone } from "../api";
import { FlaskConical, ArrowLeft, ChevronRight } from "lucide-react";

export default function DiseaseZonePage() {
  const { id } = useParams();
  const nav = useNavigate();
  const zid = parseInt(id || "0");
  const [trials, setTrials] = useState<Trial[]>([]);
  const [zone, setZone] = useState<DiseaseZone | null>(null);

  useEffect(() => {
    if (zid) {
      getZoneTrials(zid).then(setTrials);
      getDiseaseZones().then((zs) => setZone(zs.find((z) => z.id === zid) || null));
    }
  }, [zid]);

  return (
    <div>
      <div className="page-header">
        <button className="btn-text" onClick={() => nav("/")}><ArrowLeft size={16} /> 返回</button>
        <h2>{zone?.name || "病种专区"} <span className="badge">{trials.length} 项试验</span></h2>
      </div>
      <p className="muted mb-1">{zone?.description}</p>
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
      {trials.length === 0 && <p className="muted">暂无可公开试验</p>}
    </div>
  );
}
