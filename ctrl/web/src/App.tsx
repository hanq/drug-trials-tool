import React, { useState, useEffect } from "react";
import { Routes, Route, Link, useLocation, useNavigate } from "react-router-dom";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import CrawlManage from "./pages/CrawlManage";
import TrialReview from "./pages/TrialReview";
import DiseaseZones from "./pages/DiseaseZones";
import Announcements from "./pages/Announcements";
import { changePassword, updateDisplayName } from "./api";
import "./App.css";

const LS_KEY = "ctrl_admin";

export default function App() {
  const [admin, setAdmin] = useState<any>(null);
  const [ready, setReady] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [oldPwd, setOldPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [settingsMsg, setSettingsMsg] = useState("");
  const [settingsErr, setSettingsErr] = useState(false);
  const nav = useNavigate();

  useEffect(() => {
    const saved = localStorage.getItem(LS_KEY);
    if (saved) {
      const a = JSON.parse(saved);
      setAdmin(a);
      setDisplayName(a.display_name || "");
    }
    setReady(true);
  }, []);

  const handleLogin = (user: any) => {
    localStorage.setItem(LS_KEY, JSON.stringify(user));
    setAdmin(user);
    setDisplayName(user.display_name || "");
    nav("/dashboard");
  };

  const handleLogout = () => {
    localStorage.removeItem(LS_KEY);
    setAdmin(null);
    nav("/login");
  };

  const handleChangePwd = async () => {
    setSettingsMsg(""); setSettingsErr(false);
    if (!oldPwd || !newPwd) { setSettingsMsg("请填写完整"); setSettingsErr(true); return; }
    try {
      await changePassword(admin.id, oldPwd, newPwd);
      setSettingsMsg("密码修改成功"); setSettingsErr(false);
      setOldPwd(""); setNewPwd("");
    } catch (e: any) { setSettingsMsg(e.message); setSettingsErr(true); }
  };

  const handleUpdateDisplayName = async () => {
    setSettingsMsg(""); setSettingsErr(false);
    if (!displayName) { setSettingsMsg("请输入昵称"); setSettingsErr(true); return; }
    try {
      await updateDisplayName(admin.id, displayName);
      const updated = { ...admin, display_name: displayName };
      setAdmin(updated);
      localStorage.setItem(LS_KEY, JSON.stringify(updated));
      setSettingsMsg("昵称修改成功"); setSettingsErr(false);
    } catch (e: any) { setSettingsMsg(e.message); setSettingsErr(true); }
  };

  if (!ready) return <div className="login-page"><div className="spinner"></div></div>;
  if (!admin) return <Routes><Route path="*" element={<Login onLogin={handleLogin} />} /></Routes>;

  return (
    <div className="ctrl-app">
      <Sidebar admin={admin} onLogout={handleLogout} onOpenSettings={() => setShowSettings(true)} />
      <div className="main">
        <Routes>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/crawl" element={<CrawlManage />} />
          <Route path="/trials" element={<TrialReview />} />
          <Route path="/zones" element={<DiseaseZones />} />
          <Route path="/announcements" element={<Announcements />} />
          <Route path="*" element={<Dashboard />} />
        </Routes>
      </div>

      {showSettings && (
        <div className="modal-overlay" onClick={() => setShowSettings(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>账号设置</h2>

            <label>昵称</label>
            <input placeholder="输入显示昵称" value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
            <button className="btn-primary" onClick={handleUpdateDisplayName} style={{ marginBottom: "1rem" }}>保存昵称</button>

            <hr style={{ margin: "1rem 0", border: "none", borderTop: "1px solid var(--border)" }} />

            <label>旧密码</label>
            <input type="password" placeholder="输入旧密码" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)} />
            <label>新密码</label>
            <input type="password" placeholder="输入新密码" value={newPwd} onChange={(e) => setNewPwd(e.target.value)} />
            <button className="btn-primary" onClick={handleChangePwd}>修改密码</button>

            {settingsMsg && <p className="mt-1" style={{ color: settingsErr ? "red" : "var(--success)" }}>{settingsMsg}</p>}

            <button className="btn-ghost" style={{ marginTop: "1rem" }} onClick={() => setShowSettings(false)}>关闭</button>
          </div>
        </div>
      )}
    </div>
  );
}

function Sidebar({ admin, onLogout, onOpenSettings }: { admin: any; onLogout: () => void; onOpenSettings: () => void }) {
  const loc = useLocation();
  const items = [
    { path: "/dashboard", label: "Dashboard", icon: "📊" },
    { path: "/crawl", label: "爬虫管理", icon: "🕷" },
    { path: "/trials", label: "试验管理", icon: "📋" },
    { path: "/zones", label: "病种分区", icon: "📁" },
    { path: "/announcements", label: "公告管理", icon: "📢" },
  ];
  const name = admin.display_name || admin.username;
  return (
    <div className="sidebar">
      <h2>后台管理</h2>
      {items.map((i) => (
        <Link key={i.path} to={i.path} className={loc.pathname === i.path ? "active" : ""}>
          <span>{i.icon}</span> {i.label}
        </Link>
      ))}
      <div style={{ marginTop: "2rem" }}>
        <div className="muted" style={{ fontSize: "0.8rem", padding: "0 0.8rem" }}>{name}</div>
        <a href="#" onClick={(e) => { e.preventDefault(); onOpenSettings(); }} style={{ marginTop: "0.3rem", display: "block", padding: "0 0.8rem", fontSize: "0.82rem", color: "#94a3b8", cursor: "pointer" }}>账号设置</a>
        <a href="#" onClick={(e) => { e.preventDefault(); onLogout(); }} style={{ marginTop: "0.3rem" }}>退出登录</a>
      </div>
    </div>
  );
}
