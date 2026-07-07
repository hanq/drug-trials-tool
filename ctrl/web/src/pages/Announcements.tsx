import React, { useEffect, useState } from "react";
import { getAnnouncements, createAnnouncement, updateAnnouncement, deleteAnnouncement, Announcement } from "../api";

export default function Announcements() {
  const [items, setItems] = useState<Announcement[]>([]);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [editing, setEditing] = useState<Announcement | null>(null);
  const [msg, setMsg] = useState("");

  const load = () => getAnnouncements().then(setItems);
  useEffect(() => { load(); }, []);

  const handleSave = async () => {
    if (!title || !content) return;
    try {
      if (editing) {
        await updateAnnouncement(editing.id, title, content);
      } else {
        await createAnnouncement(title, content);
      }
      setTitle(""); setContent(""); setEditing(null);
      load();
    } catch (e: any) { setMsg(e.message); }
  };

  const handleEdit = (a: Announcement) => {
    setEditing(a); setTitle(a.title); setContent(a.content);
  };

  const handleDelete = async (id: number) => {
    if (!confirm("确认删除？")) return;
    try { await deleteAnnouncement(id); load(); } catch (e: any) { setMsg(e.message); }
  };

  return (
    <div>
      <div className="page-header"><h1>公告管理 <span className="badge badge-gray">{items.length}</span></h1></div>

      <div className="card">
        <h3 style={{ marginBottom: "0.8rem" }}>{editing ? "编辑公告" : "发布公告"}</h3>
        <input placeholder="标题" value={title} onChange={(e) => setTitle(e.target.value)} />
        <textarea placeholder="内容" rows={4} value={content} onChange={(e) => setContent(e.target.value)} />
        <div className="actions">
          <button className="btn-primary" onClick={handleSave}>{editing ? "保存" : "发布"}</button>
          {editing && <button className="btn-ghost" onClick={() => { setEditing(null); setTitle(""); setContent(""); }}>取消</button>}
        </div>
        {msg && <p className="muted mt-1">{msg}</p>}
      </div>

      {items.map((a) => (
        <div key={a.id} className="card" style={{ position: "relative" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
            <div>
              <strong>{a.title}</strong>
              {a.is_pinned === 1 && <span className="badge badge-yellow" style={{ marginLeft: "0.5rem" }}>置顶</span>}
              <p className="muted" style={{ marginTop: "0.3rem", whiteSpace: "pre-wrap" }}>{a.content}</p>
              <span className="muted" style={{ fontSize: "0.78rem" }}>{new Date(a.created_at).toLocaleString()}</span>
            </div>
            <div className="actions">
              <button className="btn-sm btn-ghost" onClick={() => handleEdit(a)}>编辑</button>
              <button className="btn-sm btn-danger" onClick={() => handleDelete(a.id)}>删除</button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
