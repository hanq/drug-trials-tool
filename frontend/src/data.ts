 import { Province, Institution, Investigator, Trial, ProvinceData, IndexData, OverseasData } from './types';
 
 const DATA_BASE = './data';
 
 // ── Caches ──
 let _index: IndexData | null = null;
 const _provinceCache: Record<number, ProvinceData> = {};
 let _overseas: OverseasData | null = null;
 
 // ── Index / province list ──
 export async function loadIndex(): Promise<IndexData> {
   if (_index) return _index;
   const res = await fetch(DATA_BASE + '/index.json');
   _index = await res.json();
   return _index;
 }
 
 export function getProvinces(): Province[] {
   return _index?.provinces ?? [];
 }
 export function getProvinceById(id: number): Province | undefined {
   return _index?.provinces.find(p => p.id === id);
 }
 
 // ── Per-province data ──
 export async function loadProvinceData(pid: number): Promise<ProvinceData> {
   if (_provinceCache[pid]) return _provinceCache[pid];
   const res = await fetch(DATA_BASE + `/province_${pid}.json`);
   _provinceCache[pid] = await res.json();
   return _provinceCache[pid];
 }
 
 export function getCachedProvinceData(pid: number): ProvinceData | undefined {
   return _provinceCache[pid];
 }
 
 export function getInstitutionsByProvince(pd: ProvinceData): Institution[] {
   return pd.institutions;
 }
 export function getInvestigatorsByInstitution(pd: ProvinceData, iid: number): Investigator[] {
   return pd.investigators.filter(v => v.institution_id === iid);
 }
 export function getTrialsByInvestigator(pd: ProvinceData, name: string, iid: number): Trial[] {
   const tids = pd.trial_institutions.filter(ti => ti.investigator_name === name && ti.institution_id === iid).map(ti => ti.trial_id);
   return pd.trials.filter(t => tids.includes(t.detail_id));
 }
 export function getInstitutionById(pd: ProvinceData, id: number): Institution | undefined {
   return pd.institutions.find(i => i.id === id);
 }
 export function getTrialById(pd: ProvinceData, id: string): Trial | undefined {
   return pd.trials.find(t => t.detail_id === id);
 }
 
 // ── Overseas data ──
 export async function loadOverseasData(): Promise<OverseasData> {
   if (_overseas) return _overseas;
   const res = await fetch(DATA_BASE + '/overseas.json');
   _overseas = await res.json();
   return _overseas;
 }
 export function getCachedOverseasData(): OverseasData | undefined {
   return _overseas;
 }
