import React, { useEffect, useState } from "react";
import { Announcement, getAnnouncements } from "../api";
import { ArrowLeft, Pin } from "lucide-react";
import { useNavigate } from "react-router-dom";

export default function AnnouncementsPage() {
  const [items, setItems] = useState<Announcement[]>([]);
  const nav = useNavigate();

  useEffect(() => { getAnnouncements().then(setItems); }, []);

  return (
    <div>
      <div className="page-header">
        <button className="btn-text" onClick={() => nav("/")}><ArrowLeft size={16} /> 返回</button>
        <h2>公告 <span className="badge">{items.length}</span></h2>
      </div>
      {items.map((a) => (
        <div key={a.id} className="card">
          <div className="card-body">
            <strong>{a.title}</strong>
            {a.is_pinned === 1 && <Pin size={14} className="icon" style={{ color: "#d97706" }} />}
          </div>
          <p className="muted" style={{ marginTop: "0.5rem", whiteSpace: "pre-wrap" }}>{a.content}</p>
          <span className="muted" style={{ fontSize: "0.78rem", marginTop: "0.5rem", display: "block" }}>
            {new Date(a.created_at).toLocaleDateString()}
          </span>
        </div>
      ))}
      {items.length === 0 && <p className="muted">暂无公告</p>}
    </div>
  );
}
