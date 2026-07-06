import openpyxl
import json
import os
import re
from collections import Counter, defaultdict

 # ── Config (paths relative to project root) ──
 import sys
 _script_dir = os.path.dirname(os.path.abspath(__file__))
 EXCEL_PATH = os.path.join(_script_dir, '临床试验机构信息汇总表.xlsx')
 GEO_PATH = os.path.join(_script_dir, 'frontend', 'public', 'china_map.json')
 OUT_DIR = os.path.join(_script_dir, 'frontend', 'public', 'data')

def export():
    # Build province name → (id, code) mapping from geo JSON
    geo = json.load(open(GEO_PATH, encoding='utf-8'))
    province_map = {}
    for fidx, f in enumerate(geo['features']):
        p = f['properties']
        province_map[p['name']] = {'id': fidx + 1, 'code': str(p['adcode'])}

    def normalize_province(raw):
        if not raw: return None
        raw = raw.strip()
        if raw in province_map: return raw
        lookup = {
            '\u53f0\u5317': '\u53f0\u6e7e\u7701', '\u53f0\u5357': '\u53f0\u6e7e\u7701',
            'Hsinchu': '\u53f0\u6e7e\u7701', 'Taipei': '\u53f0\u6e7e\u7701',
            '\u9999\u6e2f': '\u9999\u6e2f\u7279\u522b\u884c\u653f\u533a',
            '\u672a\u6ce8\u660e': None, '\u897f\u5b89\u5e02': '\u9655\u897f\u7701',
        }
        if raw in lookup: return lookup[raw]
        cleaned = re.sub(r'[\u7701\u5e02]$', '', raw)
        for pn in province_map:
            if cleaned == re.sub(r'[\u7701\u5e02].*$', '', pn): return pn
        return raw

    wb = openpyxl.load_workbook(EXCEL_PATH, read_only=True, data_only=True)
    ws = wb['\u9776\u70b9']
    
    rows = []
    for row in ws.iter_rows(min_row=2, values_only=True):
        target, trial_name, inst_name, pi_name, country, province, city = row[0:7]
        if not trial_name or not inst_name or not pi_name: continue
        country_str = str(country or '')
        region_type = 'china' if any(k in country_str for k in ['\u4e2d\u56fd','\u9999\u6e2f','\u53f0\u6e7e']) else 'overseas'
        rows.append({
            'target': str(target or '').strip(), 'trial_name': str(trial_name).strip(),
            'inst_name': str(inst_name).strip(), 'pi_name': str(pi_name).strip(),
            'country': country_str.strip(), 'province': normalize_province(province),
            'city': str(city or '').strip(), 'region_type': region_type,
        })

    # Build entities...
    china_rows = [r for r in rows if r['region_type'] == 'china' and r['province'] and r['province'] in province_map]
    unk = [r for r in rows if r['region_type'] == 'china' and (not r['province'] or r['province'] not in province_map)]
    if unk: print('WARNING: {} unmapped'.format(len(unk)), [r['province'] for r in unk[:5]])

    prov_name_to_id = {pname: pinfo['id'] for pname, pinfo in province_map.items()}
    provinces_list = []
    for pname, pinfo in sorted(province_map.items(), key=lambda x: x[1]['id']):
        trial_set = set(r['trial_name'] for r in china_rows if r['province'] == pname)
        provinces_list.append({'id': pinfo['id'], 'name': pname, 'code': pinfo['code'], 'trial_count': len(trial_set)})

    # Institutions
    inst_name_to_id, institutions_list, inst_counter = {}, [], 0
    for r in china_rows:
        key = (r['inst_name'], r['province'])
        if key not in inst_name_to_id:
            inst_counter += 1
            inst_name_to_id[key] = inst_counter
            institutions_list.append({'id': inst_counter, 'name': r['inst_name'],
                'province_id': prov_name_to_id[r['province']], 'city': r['city'], 'trial_count': 0})
    inst_tc = Counter()
    for r in china_rows:
        iid = inst_name_to_id.get((r['inst_name'], r['province']))
        if iid: inst_tc[iid] += 1
    for i in institutions_list: i['trial_count'] = inst_tc[i['id']]

    # PIs
    pi_key_to_id, investigators_list, pi_counter = {}, [], 0
    for r in china_rows:
        key = (r['pi_name'], r['inst_name'], r['province'])
        if key not in pi_key_to_id:
            pi_counter += 1
            pi_key_to_id[key] = pi_counter
            investigators_list.append({'id': pi_counter, 'name': r['pi_name'],
                'degree': '', 'title': '', 'phone': '', 'email': '',
                'institution_id': inst_name_to_id.get((r['inst_name'], r['province']), 0), 'trial_count': 0})
    pi_tc = Counter()
    for r in china_rows:
        pid = pi_key_to_id.get((r['pi_name'], r['inst_name'], r['province']))
        if pid: pi_tc[pid] += 1
    for i in investigators_list: i['trial_count'] = pi_tc[i['id']]

    # Trials
    trial_name_to_id, trials_list = {}, []
    trial_targets = defaultdict(set)
    for r in rows: trial_targets[r['trial_name']].add(r['target'])
    for tidx, (tname, targets) in enumerate(sorted(trial_targets.items()), 1):
        tid = 'T{:04d}'.format(tidx)
        trial_name_to_id[tname] = tid
        trials_list.append({'detail_id': tid, 'reg_no': '', 'title': tname, 'drug_name': '',
            'indication': '; '.join(sorted(targets)), 'status': '', 'applicant_name': '', 'detail_json': None})

    # Trial-Institution links
    trial_inst_list, seen_ti = [], set()
    for r in china_rows:
        trial_id = trial_name_to_id.get(r['trial_name'])
        iid = inst_name_to_id.get((r['inst_name'], r['province']))
        if not trial_id or not iid: continue
        ti_key = (trial_id, iid, r['pi_name'])
        if ti_key not in seen_ti:
            seen_ti.add(ti_key)
            trial_inst_list.append({'trial_id': trial_id, 'institution_id': iid, 'investigator_name': r['pi_name']})

    # Overseas
    overseas_rows = [r for r in rows if r['region_type'] == 'overseas']
    overseas_groups = defaultdict(list)
    for r in overseas_rows: overseas_groups[r.get('country','\u672a\u77e5') or '\u672a\u77e5'].append(r)
    
    o_provs, o_insts, o_invs, o_ti = [], [], [], []
    o_inst_map, o_inv_map = {}, {}
    o_pc = o_ic = o_vc = 0
    seen_oti = set()
    for country, grp in sorted(overseas_groups.items()):
        o_pc += 1
        ts = set(r['trial_name'] for r in grp)
        o_provs.append({'id': o_pc, 'name': country, 'code': '', 'trial_count': len(ts)})
        inst_counter_for_country = 0
        for r in grp:
            ik = (r['inst_name'], country)
            if ik not in o_inst_map:
                o_ic += 1
                inst_counter_for_country += 1
                o_inst_map[ik] = o_ic
                o_insts.append({'id': o_ic, 'name': r['inst_name'], 'province_id': o_pc, 'city': r['city'], 'trial_count': 0})
            vk = (r['pi_name'], r['inst_name'], country)
            if vk not in o_inv_map:
                o_vc += 1
                o_inv_map[vk] = o_vc
                o_invs.append({'id': o_vc, 'name': r['pi_name'], 'degree': '', 'title': '',
                    'phone': '', 'email': '', 'institution_id': o_inst_map[ik], 'trial_count': 0})
            tid = trial_name_to_id.get(r['trial_name'])
            iid = o_inst_map.get(ik)
            if tid and iid:
                tk = (tid, iid, r['pi_name'])
                if tk not in seen_oti:
                    seen_oti.add(tk)
                    o_ti.append({'trial_id': tid, 'institution_id': iid, 'investigator_name': r['pi_name']})
    
    oi_tc = Counter()
    for ti in o_ti: oi_tc[ti['institution_id']] += 1
    for i in o_insts: i['trial_count'] = oi_tc[i['id']]
    
    oiv_tc = Counter()
    seen_iv = set()
    for r in overseas_rows:
        tid = trial_name_to_id.get(r['trial_name'])
        vid = o_inv_map.get((r['pi_name'], r['inst_name'], r.get('country','') or ''))
        if tid and vid:
            k = (tid, vid)
            if k not in seen_iv:
                seen_iv.add(k); oiv_tc[vid] += 1
    for v in o_invs: v['trial_count'] = oiv_tc[v['id']]

    # Write files
    os.makedirs(OUT_DIR, exist_ok=True)
    for f in os.listdir(OUT_DIR):
        if f.startswith('province_') or f in ('index.json','overseas.json'):
            os.remove(os.path.join(OUT_DIR, f))

    json.dump({'generated_at': os.path.getmtime(EXCEL_PATH), 'provinces': provinces_list,
        'overseas': {'trial_count': len(set(r['trial_name'] for r in overseas_rows)),
            'institution_count': len(o_insts), 'investigator_count': len(o_invs), 'country_count': len(overseas_groups)}},
        open(os.path.join(OUT_DIR, 'index.json'), 'w', encoding='utf-8'), ensure_ascii=False, indent=2)
    print('index.json written')

    for p in provinces_list:
        pid = p['id']
        pi_list = [i for i in institutions_list if i['province_id'] == pid]
        iids = set(i['id'] for i in pi_list)
        pv_list = [v for v in investigators_list if v['institution_id'] in iids]
        pti = [t for t in trial_inst_list if t['institution_id'] in iids]
        ptids = set(t['trial_id'] for t in pti)
        ptr = [t for t in trials_list if t['detail_id'] in ptids]
        json.dump({'province_id': pid, 'province_name': p['name'], 'institutions': pi_list,
            'investigators': pv_list, 'trials': ptr, 'trial_institutions': pti},
            open(os.path.join(OUT_DIR, 'province_{}.json'.format(pid)), 'w', encoding='utf-8'), ensure_ascii=False, indent=2)

    json.dump({'provinces': o_provs, 'institutions': o_insts, 'investigators': o_invs,
        'trials': trials_list, 'trial_institutions': o_ti},
        open(os.path.join(OUT_DIR, 'overseas.json'), 'w', encoding='utf-8'), ensure_ascii=False, indent=2)
    print('overseas.json written')
    print('Done. {} provinces, {} inst, {} inv, {} trials'.format(len(provinces_list), len(institutions_list)+len(o_insts), len(investigators_list)+len(o_invs), len(trials_list)))

if __name__ == '__main__':
    export()
