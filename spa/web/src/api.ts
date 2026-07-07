const API = "http://localhost:8080/api/spa";

interface Opts { method?: string; body?: any }

async function req(path: string, opts: Opts = {}) {
  const res = await fetch(API + path, {
    method: opts.method || "GET",
    headers: opts.body ? { "Content-Type": "application/json" } : undefined,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });
  const data = await res.json();
  if (!data.success) throw new Error(data.error || "request failed");
  return data.data;
}

export interface Province { id: number; name: string; code: string; trial_count: number }
export interface Institution { id: number; name: string; province_id: number; city: string; trial_count: number }
export interface Investigator { id: number; name: string; degree?: string; title?: string; institution_id: number; trial_count: number }
export interface Trial { detail_id: string; reg_no?: string; title: string; drug_name?: string; indication: string; status?: string; applicant_name?: string; detail_json?: any; disease_zone_id?: number }
export interface DiseaseZone { id: number; name: string; keyword: string; description: string }
export interface Announcement { id: number; title: string; content: string; is_pinned: number; created_at: string }
export interface SearchResult { trials: Trial[]; total: number; page: number; page_size: number }

export function getProvinces() { return req("/provinces") as Promise<Province[]> }
export function getInstitutions(pid: number) { return req("/provinces/" + pid) as Promise<Institution[]> }
export function getInvestigators(iid: number) { return req("/institutions/" + iid) as Promise<Investigator[]> }
export function getTrials(name: string, iid: number) { return req("/investigators?" + new URLSearchParams({ name, institution_id: String(iid) })) as Promise<Trial[]> }
export function getTrial(id: string) { return req("/trials/" + id) as Promise<Trial> }
export function search(q: string, page?: number) { return req("/search?" + new URLSearchParams({ q, page_size: "20", page: String(page || 1) })) as Promise<SearchResult> }
export function getDiseaseZones() { return req("/disease-zones") as Promise<DiseaseZone[]> }
export function getZoneTrials(zid: number) { return req("/disease-zones/" + zid) as Promise<Trial[]> }
export function getAnnouncements() { return req("/announcements") as Promise<Announcement[]> }
