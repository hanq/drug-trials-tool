import React, { lazy, Suspense } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import Header from "./components/Header";
import { ZoneProvider } from "./ZoneContext";

const Home = lazy(() => import("./pages/Home"));
const Province = lazy(() => import("./pages/Province"));
const Institution = lazy(() => import("./pages/Institution"));
const Investigator = lazy(() => import("./pages/Investigator"));
const TrialDetail = lazy(() => import("./pages/TrialDetail"));
const Search = lazy(() => import("./pages/Search"));
const DiseaseZone = lazy(() => import("./pages/DiseaseZone"));
const Announcements = lazy(() => import("./pages/Announcements"));

export default function App() {
  return (
    <div className="spa-app">
      <ZoneProvider>
      <Header />
      <main className="main-content">
        <Suspense fallback={<div className="app-loading"><div className="spinner"></div></div>}>
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
        </Suspense>
      </main>
      </ZoneProvider>
    </div>
  );
}
