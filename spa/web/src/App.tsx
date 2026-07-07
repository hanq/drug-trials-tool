import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';

export default function App() {
  return (
    <div className="spa-app">
      <React.Suspense fallback={<div>加载中...</div>}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/province/:id" element={<Province />} />
          <Route path="/institution/:id" element={<Institution />} />
          <Route path="/investigator" element={<Investigator />} />
          <Route path="/trial/:id" element={<TrialDetail />} />
          <Route path="/search" element={<Search />} />
          <Route path="/zone/:id" element={<DiseaseZone />} />
          <Route path="/announcements" element={<Announcements />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </React.Suspense>
    </div>
  );
}

const Home = React.lazy(() => import("./pages/Home"));
const Province = React.lazy(() => import("./pages/Province"));
const Institution = React.lazy(() => import("./pages/Institution"));
const Investigator = React.lazy(() => import("./pages/Investigator"));
const TrialDetail = React.lazy(() => import("./pages/TrialDetail"));
const Search = React.lazy(() => import("./pages/Search"));
const DiseaseZone = React.lazy(() => import("./pages/DiseaseZone"));
const AnnouncementsLazy = React.lazy(() => import("./pages/Announcements"));
