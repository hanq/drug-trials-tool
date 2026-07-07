 import React, { useState, useEffect } from 'react';
 import { useNavigate, useLocation, useParams } from 'react-router-dom';
 import { Province, Trial, ProvinceData } from './types';
 import { getProvinceById, loadProvinceData, getCachedProvinceData, getInstitutionsByProvince, getInvestigatorsByInstitution, getTrialsByInvestigator } from './data';
 import ChinaMap from './components/ChinaMap';
 import { MapPin, List, Building2, User, FlaskConical, ChevronRight, ArrowLeft } from 'lucide-react';
 
 // ── Hook: load province data with cache ──
 function useProvinceData(pid: number): { loading: boolean; provinceData: ProvinceData | null } {
   const [loading, setLoading] = useState(true);
   const [pd, setPd] = useState<ProvinceData | null>(null);
 
   useEffect(() => {
     const cached = getCachedProvinceData(pid);
     if (cached) { setPd(cached); setLoading(false); return; }
     setLoading(true);
     loadProvinceData(pid).then(d => { setPd(d); setLoading(false); });
   }, [pid]);
 
   return { loading, provinceData: pd };
 }
 
 // ── HomePage ──
 export function HomePage({ provinces, searchQuery }: { provinces: Province[]; searchQuery: string }) {
   const [view, setView] = useState<'map' | 'list'>('list');
   const nav = useNavigate();
   const loc = useLocation();
   const q = new URLSearchParams(loc.search).get('q') || searchQuery;
 
   const handleProvinceClick = (name: string) => {
     const p = provinces.find(pr => name.includes(pr.name.replace(/[市省]/g,'')));
     if (p) nav('/province/' + p.id);
   };
 
   const totalTrials = provinces.reduce((s, p) => s + p.trial_count, 0);
 
   return (
     <div>
       <div className="page-header">
         <h2><span className="badge">{provinces.length} 省 / {totalTrials} 项试验</span></h2>
         <div className="view-toggle">
           <button className={'btn-toggle ' + (view==='map'?'active':'')} onClick={()=>setView('map')}><MapPin size={16} /> 地图</button>
           <button className={'btn-toggle ' + (view==='list'?'active':'')} onClick={()=>setView('list')}><List size={16} /> 列表</button>
         </div>
       </div>
       {!q && view === 'map' && <ChinaMap provinces={provinces} onProvinceClick={handleProvinceClick} />}
       {!q && view === 'list' && (
         <div className="province-grid">
           {provinces.map(p => (
             <div key={p.id} className="card card-click" onClick={() => nav('/province/'+p.id)}>
               <div className="card-body"><strong>{p.name}</strong><span className="badge">{p.trial_count}</span></div>
             </div>
           ))}
         </div>
       )}
     </div>
   );
 }
 
 // ── InstitutionPage ──
 export function InstitutionPage() {
   const { id } = useParams();
   const nav = useNavigate();
   const pid = parseInt(id || '0');
   const { loading, provinceData } = useProvinceData(pid);
   const province = getProvinceById(pid);
 
   if (loading) return <div className="app-loading"><div className="spinner"></div><p>加载中...</p></div>;
   if (!provinceData) return <div className="app-loading"><p>数据加载失败</p></div>;
 
   const institutions = getInstitutionsByProvince(provinceData);
   return (
     <div>
       <div className="page-header">
         <button className="btn-text" onClick={() => nav('/')}><ArrowLeft size={16} /> 返回</button>
         <h2>{province?.name || '省份'} <span className="badge">{institutions.length} 家机构</span></h2>
       </div>
       <div className="inst-grid">{institutions.map(inst => (
         <div key={inst.id} className="card card-click" onClick={() => nav('/institution/'+inst.id+'?pid='+pid)}>
           <div className="card-body">
             <Building2 size={18} className="icon" />
             <div><strong>{inst.name}</strong><span className="muted">{inst.city}</span></div>
             <span className="badge">{inst.trial_count} 项</span>
           </div>
         </div>
       ))}</div>
     </div>
   );
 }
 
 // ── InvestigatorPage ──
 export function InvestigatorPage() {
   const { id } = useParams();
   const nav = useNavigate();
   const loc = useLocation();
   const params = new URLSearchParams(loc.search);
   const pid = parseInt(params.get('pid') || '0');
   const iid = parseInt(id || '0');
   const { loading, provinceData } = useProvinceData(pid);
 
   if (loading) return <div className="app-loading"><div className="spinner"></div><p>加载中...</p></div>;
 
   const inst = provinceData?.institutions.find(i => i.id === iid);
   const investigators = provinceData ? getInvestigatorsByInstitution(provinceData, iid) : [];
 
   return (
     <div>
       <div className="page-header">
         <button className="btn-text" onClick={() => nav(-1)}><ArrowLeft size={16} /> 返回</button>
         <h2>{inst?.name || '机构'} <span className="badge">{investigators.length} 位研究者</span></h2>
       </div>
       <div className="inv-grid">{investigators.map(inv => (
         <div key={inv.id} className="card card-click" onClick={() => nav('/investigator?name='+encodeURIComponent(inv.name)+'&iid='+inv.institution_id+'&pid='+pid)}>
           <div className="card-body">
             <User size={18} className="icon" />
             <div><strong>{inv.name}</strong><span className="muted">{inv.title || ''} {inv.degree || ''}</span></div>
             <span className="badge">{inv.trial_count} 项</span>
           </div>
         </div>
       ))}</div>
     </div>
   );
 }
 
 // ── TrialPage ──
 export function TrialPage() {
   const nav = useNavigate();
   const loc = useLocation();
   const params = new URLSearchParams(loc.search);
   const name = params.get('name') || '';
   const iid = parseInt(params.get('iid') || '0');
   const pid = parseInt(params.get('pid') || '0');
   const [selected, setSelected] = useState<Trial | null>(null);
   const { loading, provinceData } = useProvinceData(pid);
 
   if (loading) return <div className="app-loading"><div className="spinner"></div><p>加载中...</p></div>;
 
   const inst = provinceData?.institutions.find(i => i.id === iid);
   const trials = provinceData ? getTrialsByInvestigator(provinceData, name, iid) : [];
 
   return (
     <div>
       <div className="page-header">
         <button className="btn-text" onClick={() => nav(-1)}><ArrowLeft size={16} /> 返回</button>
         <h2>{name} <span className="muted">{inst?.name}</span> <span className="badge">{trials.length} 项试验</span></h2>
       </div>
       <TrialList trials={trials} onSelect={setSelected} />
       {selected && <TrialModal trial={selected} onClose={() => setSelected(null)} />}
     </div>
   );
 }
 
 // ── Shared Components ──
 function TrialList({ trials, onSelect }: { trials: Trial[]; onSelect?: (t: Trial) => void }) {
   return (<div className="trial-list">{trials.map(t => (
     <div key={t.detail_id} className="card card-click" onClick={() => onSelect?.(t)}>
       <div className="card-body">
         <FlaskConical size={18} className="icon" />
         <div className="trial-info">
           <strong>{t.title}</strong>
           <span className="muted">{t.indication?.slice(0,40)}</span>
         </div>
         <ChevronRight size={16} className="chevron" />
       </div>
     </div>
   ))}</div>);
 }
 
 function TrialModal({ trial, onClose }: { trial: Trial; onClose: () => void }) {
   return (<div className="modal-overlay" onClick={onClose}>
     <div className="modal" onClick={e => e.stopPropagation()}>
       <div className="modal-header"><h3>{trial.title}</h3><button className="btn-text" onClick={onClose}>✕</button></div>
       <div className="modal-body">
         <table className="detail-table"><tbody>
           <tr><th>靶点/适应症</th><td>{trial.indication}</td></tr>
         </tbody></table>
       </div>
     </div>
   </div>);
 }
 
 // ── AdminPage ──
 export function AdminPage() {
   return (<div>
     <div className="page-header"><h2>后台管理</h2></div>
     <div className="admin-panel">
       <div className="card"><div className="card-body">
         <h3>数据源</h3>
         <p className="muted">数据来源于《临床试验机构信息汇总表.xlsx》，按省份拆分加载</p>
       </div></div>
     </div>
   </div>);
 }
