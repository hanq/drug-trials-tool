 export interface Province {
   id: number; name: string; code: string; trial_count: number;
 }
 export interface Institution {
   id: number; name: string; province_id: number; city: string; trial_count: number;
 }
 export interface Investigator {
   id: number; name: string; degree?: string; title?: string; phone?: string; email?: string; institution_id: number; trial_count: number;
 }
 export interface Trial {
   detail_id: string; reg_no?: string; title: string; drug_name?: string; indication: string; status?: string; applicant_name?: string; detail_json?: any;
 }
 export interface TrialInstitution {
   trial_id: string; institution_id: number; investigator_name: string;
 }
 export interface ExportData {
   generated_at: string; provinces: Province[]; institutions: Institution[]; investigators: Investigator[]; trials: Trial[]; trial_institutions: TrialInstitution[];
 }
 export interface ProvinceData {
   province_id: number;
   province_name: string;
   institutions: Institution[];
   investigators: Investigator[];
   trials: Trial[];
   trial_institutions: TrialInstitution[];
 }
 export interface IndexData {
   generated_at: number;
   provinces: Province[];
   overseas: { trial_count: number; institution_count: number; investigator_count: number; country_count: number };
 }
 export interface OverseasData {
   provinces: Province[];
   institutions: Institution[];
   investigators: Investigator[];
   trials: Trial[];
   trial_institutions: TrialInstitution[];
 }
