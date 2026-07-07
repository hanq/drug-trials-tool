import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Search, Database } from "lucide-react";

export default function Header({ onSearch }: { onSearch?: (q: string) => void }) {
  const [q, setQ] = useState("");
  const nav = useNavigate();

  const doSearch = () => {
    if (q.trim()) { nav("/search?q=" + encodeURIComponent(q)); onSearch?.(q); }
  };

  return (
    <header className="spa-header">
      <div className="header-inner">
        <div className="header-logo" onClick={() => nav("/")} style={{ cursor: "pointer" }}>
          <Database size={22} /> 就近找临床
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
