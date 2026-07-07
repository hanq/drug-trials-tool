import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';

const pages = ["Login","Dashboard","CrawlManage","TrialReview","DiseaseZones","Announcements"];
const LazyPage = (name: string) => React.lazy(() => import(`./pages/${name}`));

export default function App() {
  return (
    <div className="ctrl-app">
      <React.Suspense fallback={<div>Loading...</div>}>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/crawl" element={<CrawlManage />} />
          <Route path="/trials" element={<TrialReview />} />
          <Route path="/zones" element={<DiseaseZones />} />
          <Route path="/announcements" element={<Announcements />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </React.Suspense>
    </div>
  );
}
