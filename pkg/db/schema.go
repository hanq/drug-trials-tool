package db

import "database/sql"

const Schema = `
-- 管理员
CREATE TABLE IF NOT EXISTS admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 病种分区
CREATE TABLE IF NOT EXISTS disease_zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    keyword TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 省
CREATE TABLE IF NOT EXISTS provinces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    code TEXT
);

-- 机构
CREATE TABLE IF NOT EXISTS institutions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    province_id INTEGER REFERENCES provinces(id),
    city TEXT
);

-- 主要研究者
CREATE TABLE IF NOT EXISTS investigators (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    degree TEXT, title TEXT, phone TEXT,
    email TEXT, address TEXT, zip_code TEXT,
    institution_id INTEGER REFERENCES institutions(id),
    UNIQUE(name, institution_id)
);

-- 试验
CREATE TABLE IF NOT EXISTS trials (
    detail_id TEXT PRIMARY KEY,
    reg_no TEXT, title TEXT, drug_name TEXT,
    indication TEXT, status TEXT, applicant_name TEXT,
    keyword TEXT,
    detail_json TEXT,
    crawl_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_hash TEXT,
    published INTEGER DEFAULT 0,
    published_at TIMESTAMP,
    published_by INTEGER REFERENCES admin_users(id),
    disease_zone_id INTEGER REFERENCES disease_zones(id)
);

-- 试验-机构关联
CREATE TABLE IF NOT EXISTS trial_institutions (
    trial_id TEXT REFERENCES trials(detail_id),
    institution_id INTEGER REFERENCES institutions(id),
    investigator_name TEXT,
    PRIMARY KEY(trial_id, institution_id, investigator_name)
);

-- 公告
CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_pinned INTEGER DEFAULT 0,
    published INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES admin_users(id)
);

-- 爬虫日志
CREATE TABLE IF NOT EXISTS crawl_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    keyword TEXT NOT NULL,
    pages INTEGER DEFAULT 0,
    disease_zone_id INTEGER REFERENCES disease_zones(id),
    found INTEGER DEFAULT 0,
    new_items INTEGER DEFAULT 0,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    status TEXT DEFAULT 'running'
);
`

func MigrateDB(db *sql.DB) error {
    _, err := db.Exec(Schema)
    return err
}
