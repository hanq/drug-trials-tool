import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Institution, getInstitutions, getProvinces, Province } from "../api";
import { Building2, ArrowLeft } from "lucide-react";

export default function ProvincePage() {
  const { id } = useParams();
  const nav = useNavigate();
  const pid = parseInt(id || "0");
  const [insts, setInsts] = useState<Institution[]>([]);
  const [province, setProv] = useState<Province | null>(null);

  useEffect(() => {
    getInstitutions(pid, selectedZone?.id).then(setInsts);
    getProvinces(selectedZone?.id).then((ps) => setProv(ps.find((p) => p.id === pid) || null));
  }, [pid, selectedZone?.id]);

  return (
    <div>
      <div className="page-header">
        <button className="btn-text" onClick={() => nav("/")}><ArrowLeft size={16} /> 返回</button>
        <h2>{province?.name || "省份"} <span className="badge">{insts.length} 家机构</span></h2>
      </div>
      <div className="inst-grid">
        {insts.map((inst) => (
          <div key={inst.id} className="card card-click" onClick={() => nav("/institution/" + inst.id + "?pid=" + pid)}>
            <div className="card-body">
              <Building2 size={18} className="icon" />
              <div><strong>{inst.name}</strong><span className="muted">{inst.city}</span></div>
              <span className="badge">{inst.trial_count} 项</span>
            </div>
          </div>
        ))}
        {insts.length === 0 && <p className="muted">暂无数据</p>}
      </div>
    </div>
  );
}
import { useZone } from "../ZoneContext";
  const { selectedZone } = useZone();
