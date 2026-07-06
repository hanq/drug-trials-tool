 import React, { useState, useEffect } from 'react';
 import { Routes, Route } from 'react-router-dom';
 import { Province } from './types';
 import { loadIndex } from './data';
 import Header from './components/Header';
 import { HomePage, InstitutionPage, InvestigatorPage, TrialPage, AdminPage } from './pages_all';
 
 function App() {
   const [provinces, setProvinces] = useState<Province[] | null>(null);
   const [loading, setLoading] = useState(true);
   const [error, setError] = useState('');
   const [searchQuery, setSearchQuery] = useState('');
 
   useEffect(() => {
     loadIndex().then(idx => {
       setProvinces(idx.provinces);
       setLoading(false);
     }).catch(e => {
       setError('数据加载失败: ' + e.message);
       setLoading(false);
     });
   }, []);
 
   if (loading) return <div className="app-loading"><div className="spinner"></div><p>加载数据中...</p></div>;
   if (error) return <div className="app-loading"><p>{error}</p></div>;
   if (!provinces) return <div className="app-loading"><p>数据加载失败</p></div>;
 
   return (
     <div className="app">
       <Header onSearch={setSearchQuery} />
       <main className="main-content">
         <Routes>
           <Route path="/" element={<HomePage provinces={provinces} searchQuery={searchQuery} />} />
           <Route path="/province/:id" element={<InstitutionPage />} />
           <Route path="/institution/:id" element={<InvestigatorPage />} />
           <Route path="/investigator" element={<TrialPage />} />
           <Route path="/overseas" element={<InstitutionPage />} />
           <Route path="/admin" element={<AdminPage />} />
         </Routes>
       </main>
     </div>
   );
 }
 
 export default App;
