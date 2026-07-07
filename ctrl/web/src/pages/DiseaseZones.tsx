import React, { useEffect, useState } from "react";
import { getDiseaseZones, createDiseaseZone, deleteDiseaseZone, DiseaseZone } from "../api";

export default function DiseaseZones() {
  const [zones, setZones] = useState<DiseaseZone[]>([]);
  const [name, setName] = useState("");
  const [keyword, setKeyword] = useState("");
  const [desc, setDesc] = useState("");
  const [msg, setMsg] = useState("");

  const load = () => getDiseaseZones().then(setZones);
  useEffect(() => { load(); }, []);

  const handleCreate = async () => {
    if (!name || !keyword) return;
    try {
      await createDiseaseZone(name, keyword, desc);
      setName(""); setKeyword(""); setDesc("");
      load();
    } catch (e: any) { setMsg(e.message); }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("确认删除？")) return;
    try { await deleteDiseaseZone(id); load(); } catch (e: any) { setMsg(e.message); }
  };

  return (
    <div>
      <div className="page-header"><h1>病种分区 <span className="badge badge-gray">{zones.length}</span></h1></div>

      <div className="card">
        <h3 style={{ marginBottom: "0.8rem" }}>新建分区</h3>
        <div className="form-row">
          <div><label>名称</label><input placeholder="如: 胰腺癌" value={name} onChange={(e) => setName(e.target.value)} /></div>
          <div><label>搜索关键词</label><input placeholder="如: 胰腺" value={keyword} onChange={(e) => setKeyword(e.target.value)} /></div>
          <div><label>描述</label><input placeholder="可选" value={desc} onChange={(e) => setDesc(e.target.value)} /></div>
          <div style={{ display: "flex", alignItems: "flex-end" }}>
            <button className="btn-primary" onClick={handleCreate}>创建</button>
          </div>
        </div>
        {msg && <p className="muted mt-1">{msg}</p>}
      </div>

      <table>
        <thead><tr><th>ID</th><th>名称</th><th>关键词</th><th>描述</th><th>操作</th></tr></thead>
        <tbody>
          {zones.map((z) => (
            <tr key={z.id}>
              <td>{z.id}</td>
              <td><strong>{z.name}</strong></td>
              <td><span className="badge badge-gray">{z.keyword}</span></td>
              <td className="muted">{z.description}</td>
              <td><button className="btn-sm btn-danger" onClick={() => handleDelete(z.id)}>删除</button></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
