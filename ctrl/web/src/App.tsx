import React, { useState, useEffect } from "react";
import { Routes, Route, Link, useLocation, useNavigate } from "react-router-dom";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import CrawlManage from "./pages/CrawlManage";
import TrialReview from "./pages/TrialReview";
import DiseaseZones from "./pages/DiseaseZones";
import Announcements from "./pages/Announcements";
import "./App.css";

const LS_KEY = "ctrl_admin";

export default function App() {
  const [admin, setAdmin] = useState<any>(null);
  const [ready, setReady] = useState(false);
  const nav = useNavigate();

  useEffect(() => {
    const saved = localStorage.getItem(LS_KEY);
    if (saved) setAdmin(JSON.parse(saved));
    setReady(true);
  }, []);

  const handleLogin = (user: any) => {
    localStorage.setItem(LS_KEY, JSON.stringify(user));
    setAdmin(user);
    nav("/dashboard");
  };

  const handleLogout = () => {
    localStorage.removeItem(LS_KEY);
    setAdmin(null);
    nav("/login");
  };

  if (!ready) return <div className="login-page"><div className="spinner"></div></div>;
  if (!admin) return <Routes><Route path="*" element={<Login onLogin={handleLogin} />} /></Routes>;

  return (
    <div className="ctrl-app">
      <Sidebar admin={admin} onLogout={handleLogout} />
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
    </div>
  );
}

function Sidebar({ admin, onLogout }: { admin: any; onLogout: () => void }) {
  const loc = useLocation();
  const items = [
    { path: "/dashboard", label: "Dashboard", icon: "📊" },
    { path: "/crawl", label: "爬虫管理", icon: "🕷" },
    { path: "/trials", label: "试验审核", icon: "📋" },
    { path: "/zones", label: "病种分区", icon: "📁" },
    { path: "/announcements", label: "公告管理", icon: "📢" },
  ];
  return (
    <div className="sidebar">
      <h2>后台管理</h2>
      {items.map((i) => (
        <Link key={i.path} to={i.path} className={loc.pathname === i.path ? "active" : ""}>
          <span>{i.icon}</span> {i.label}
        </Link>
      ))}
      <div style={{ marginTop: "2rem" }}>
        <div className="muted" style={{ fontSize: "0.8rem", padding: "0 0.8rem" }}>{admin.username}</div>
        <a href="#" onClick={(e) => { e.preventDefault(); onLogout(); }} style={{ marginTop: "0.3rem" }}>退出登录</a>
      </div>
    </div>
  );
}
