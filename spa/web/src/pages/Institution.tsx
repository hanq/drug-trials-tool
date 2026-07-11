import React, { useEffect, useState } from "react";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { Investigator, Institution, getInvestigators, getInstitutions } from "../api";
import { useZone } from "../ZoneContext";
import { User, ArrowLeft, Building2 } from "lucide-react";

export default function InstitutionPage() {
  const { id } = useParams();
  const [params] = useSearchParams();
  const nav = useNavigate();
  const iid = parseInt(id || "0");
  const pid = parseInt(params.get("pid") || "0");
  const { selectedZone } = useZone();
  const [invs, setInvs] = useState<Investigator[]>([]);
  const [inst, setInst] = useState<Institution | null>(null);

  useEffect(() => {
    getInvestigators(iid, selectedZone?.id).then(setInvs);
    if (pid) getInstitutions(pid, selectedZone?.id).then((is) => setInst(is.find((i) => i.id === iid) || null));
  }, [iid, pid, selectedZone?.id]);

  return (
    <div>
      <div className="page-header">
        <button className="btn-text" onClick={() => nav(-1)}><ArrowLeft size={16} /> 返回</button>
        <h2>{inst?.name || "机构"} <span className="badge">{invs.length} 位研究者</span></h2>
      </div>
      <div className="inv-grid">
        {invs.map((v) => (
          <div key={v.id} className="card card-click" onClick={() => nav("/investigator?name=" + encodeURIComponent(v.name) + "&iid=" + v.institution_id + "&pid=" + pid)}>
            <div className="card-body">
              <User size={18} className="icon" />
              <div><strong>{v.name}</strong><span className="muted">{v.title || ""} {v.degree || ""}</span></div>
              <span className="badge">{v.trial_count} 项</span>
            </div>
          </div>
        ))}
        {invs.length === 0 && <p className="muted">暂无研究者</p>}
      </div>
    </div>
  );
}
