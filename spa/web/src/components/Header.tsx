import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Search, Database, ChevronDown } from "lucide-react";
import { useZone } from "../ZoneContext";

export default function Header() {
  const [q, setQ] = useState("");
  const nav = useNavigate();
  const { zones, selectedZone, setSelectedZoneId, loading } = useZone();

  const handleZoneChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value;
    const zid = val ? parseInt(val, 10) : null;
    setSelectedZoneId(zid);
  };

  const doSearch = () => {
    if (q.trim()) {
    const params = new URLSearchParams();
    params.set("q", q.trim());
    if (selectedZone) params.set("zone_id", String(selectedZone.id));
    nav("/search?" + params.toString());
  }
  };

  return (
    <header className="spa-header">
      <div className="header-inner">
        <div className="header-logo" onClick={() => nav("/")} style={{ cursor: "pointer" }}>
          <Database size={22} /> 就近找临床
        </div>
        <div className="zone-selector">
          <select
            value={selectedZone?.id || ""}
            onChange={handleZoneChange}
            disabled={loading}
          >
            <option value="">全部病种</option>
            {zones.map((z) => (
              <option key={z.id} value={z.id}>{z.name}</option>
            ))}
          </select>
          <ChevronDown size={14} className="zone-chevron" />
        </div>
        <div className="header-search">
          <input
            placeholder="搜索试验（标题/药物/适应症/登记号）" value={q}
            onChange={(e) => setQ(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && doSearch()}
          />
          <Search size={18} className="search-icon" onClick={doSearch} />
        </div>
      </div>
    </header>
  );
}
