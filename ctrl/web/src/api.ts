// API client for ctrl (admin backend)
const API = "http://localhost:8081/api/ctrl";

interface ReqOpts {
  method?: string;
  body?: any;
}

async function request(path: string, opts: ReqOpts = {}) {
  const res = await fetch(API + path, {
    method: opts.method || "GET",
    headers: opts.body ? { "Content-Type": "application/json" } : undefined,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });
  const data = await res.json();
  if (!data.success) throw new Error(data.error || "request failed");
  return data.data;
}

export interface DashboardStats {
  total_trials: number; published_trials: number; pending_trials: number;
  total_provinces: number; total_institutions: number; total_investigators: number;
  total_zones: number; recent_crawls: CrawlLog[];
}

export interface Trial { detail_id: string; reg_no: string; title: string; drug_name: string; indication: string; status: string; applicant_name: string; keyword: string; crawl_time: string; published: number; disease_zone_id: number; }
export interface DiseaseZone { id: number; name: string; keyword: string; description: string; }
export interface Announcement { id: number; title: string; content: string; is_pinned: number; published: number; created_at: string; }
export interface CrawlLog { id: number; keyword: string; pages: number; found: number; new_items: number; start_time: string; end_time: string; status: string; }

export function login(username: string, password: string) { return request("/login", { method: "POST", body: { username, password } }); }
export function getDashboard() { return request("/dashboard") as Promise<DashboardStats>; }
export function getTrials(zoneId?: number, keyword?: string) {
  const p = new URLSearchParams();
  if (zoneId) p.set("zone_id", String(zoneId));
  if (keyword) p.set("keyword", keyword);
  return request("/trials?" + p.toString()) as Promise<Trial[]>;
}
export function publishTrial(detailId: string) { return request("/trials/" + detailId + "/publish", { method: "POST" }); }
export function batchPublish(zoneId?: number, keyword?: string) { return request("/trials", { method: "POST", body: { zone_id: zoneId, keyword } }); }
export function getDiseaseZones() { return request("/disease-zones") as Promise<DiseaseZone[]>; }
export function createDiseaseZone(name: string, keyword: string, desc: string) { return request("/disease-zones", { method: "POST", body: { name, keyword, description: desc } }); }
export function deleteDiseaseZone(id: number) { return request("/disease-zones/" + id, { method: "DELETE" }); }
export function crawlStart(keyword: string, pages: number, zoneId: number) { return request("/crawl/start", { method: "POST", body: { keyword, pages, zone_id: zoneId } }); }
export function crawlStatus() { return request("/crawl/status") as Promise<{crawling: boolean}>; }
export function getCrawlLogs() { return request("/crawl/logs") as Promise<CrawlLog[]>; }
export function getAnnouncements() { return request("/announcements") as Promise<Announcement[]>; }
export function createAnnouncement(title: string, content: string) { return request("/announcements", { method: "POST", body: { title, content } }); }
export function updateAnnouncement(id: number, title: string, content: string) { return request("/announcements/" + id, { method: "PUT", body: { title, content } }); }
export function deleteAnnouncement(id: number) { return request("/announcements/" + id, { method: "DELETE" }); }
