import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Province, getProvinces } from "../api";
import ChinaMap from "../components/ChinaMap";
import { MapPin, List, Building2 } from "lucide-react";

export default function Home() {
  const [provinces, setProv] = useState<Province[]>([]);
  const [view, setView] = useState<"map" | "list">("list");
  const [params] = useSearchParams();
  const nav = useNavigate();
  const q = params.get("q") || "";

  useEffect(() => { getProvinces().then(setProv); }, []);

  const filtered = q
    ? provinces.filter((p) => p.name.includes(q) || p.name.includes(q.replace(/[省市]/g, "")))
    : provinces;

  const totalTrials = provinces.reduce((s, p) => s + p.trial_count, 0);

  const handleProvinceClick = (name: string) => {
    const p = provinces.find((pr) => name.includes(pr.name.replace(/[市省]/g, "")));
    if (p) nav("/province/" + p.id);
  };

  return (
    <div>
      <div className="page-header">
        <h2><span className="badge">{provinces.length} 省 / {totalTrials} 项试验</span></h2>
        <div className="view-toggle">
          <button className={"btn-toggle " + (view === "map" ? "active" : "")} onClick={() => setView("map")}>
            <MapPin size={16} /> 地图
          </button>
          <button className={"btn-toggle " + (view === "list" ? "active" : "")} onClick={() => setView("list")}>
            <List size={16} /> 列表
          </button>
        </div>
      </div>

      {!q && view === "map" && <ChinaMap provinces={provinces} onProvinceClick={handleProvinceClick} />}

      <div className="province-grid">
        {filtered.map((p) => (
          <div key={p.id} className="card card-click" onClick={() => nav("/province/" + p.id)}>
            <div className="card-body">
              <Building2 size={18} className="icon" />
              <strong>{p.name}</strong>
              <span className="badge">{p.trial_count} 项</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
