import React, { useState } from "react";
import { login } from "../api";

interface Props { onLogin: (user: any) => void; }

export default function LoginPage({ onLogin }: Props) {
  const [u, setU] = useState("");
  const [p, setP] = useState("");
  const [err, setErr] = useState("");

  const handle = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr("");
    try {
      const user = await login(u, p);
      onLogin(user);
    } catch (ex: any) {
      setErr(ex.message);
    }
  };

  return (
    <div className="login-page">
      <form className="login-card" onSubmit={handle}>
        <h1>后台管理登录</h1>
        {err && <p style={{ color: "red", fontSize: "0.85rem", marginBottom: "0.8rem" }}>{err}</p>}
        <input placeholder="用户名" value={u} onChange={(e) => setU(e.target.value)} required />
        <input type="password" placeholder="密码" value={p} onChange={(e) => setP(e.target.value)} required />
        <button type="submit" className="btn-primary" style={{ width: "100%", padding: "0.6rem" }}>登录</button>
      </form>
    </div>
  );
}
