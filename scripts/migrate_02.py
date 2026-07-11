#!/usr/bin/env python3
"""
0.2 数据库迁移脚本

将 0.1 架构的 trials.db 升级到 0.2 架构:
  1. 创建新表 (admin_users, disease_zones, announcements, crawl_logs)
  2. ALTER TABLE trials 补充缺失列
  3. 设置旧数据 published=1
  4. 插入初始管理员
  5. 清理 crawl_sessions → crawl_logs 过渡
"""

import sqlite3
import sys
from pathlib import Path


def migrate(db_path: str) -> None:
    db = sqlite3.connect(db_path)
    db.execute("PRAGMA foreign_keys=OFF")
    db.execute("PRAGMA journal_mode=WAL")

    print("=== 0.2 数据库迁移 ===")
    print(f"数据库: {db_path}")

    # ---------------------------------------------------------------
    # 1. 创建新表 (IF NOT EXISTS)
    # ---------------------------------------------------------------
    print("\n[1/5] 创建新表...")
    new_tables = """
        CREATE TABLE IF NOT EXISTS admin_users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS disease_zones (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            keyword TEXT NOT NULL,
            description TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
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
    """
    db.executescript(new_tables)
    created = []
    for tbl in ["admin_users", "disease_zones", "announcements", "crawl_logs"]:
        cnt = db.execute(f"SELECT COUNT(*) FROM {tbl}").fetchone()[0]
        created.append(f"  {tbl}: {cnt} rows")
    print("\n".join(created))

    # ---------------------------------------------------------------
    # 2. ALTER TABLE trials 补充缺失列
    # ---------------------------------------------------------------
    print("\n[2/5] 检查 trials 表缺失列...")
    cols = {row[1] for row in db.execute("PRAGMA table_info(trials)").fetchall()}
    missing_cols = []

    if "published" not in cols:
        db.execute("ALTER TABLE trials ADD COLUMN published INTEGER DEFAULT 0")
        missing_cols.append("published")
    if "published_at" not in cols:
        db.execute("ALTER TABLE trials ADD COLUMN published_at TIMESTAMP")
        missing_cols.append("published_at")
    if "published_by" not in cols:
        db.execute("ALTER TABLE trials ADD COLUMN published_by INTEGER REFERENCES admin_users(id)")
        missing_cols.append("published_by")
    if "disease_zone_id" not in cols:
        db.execute("ALTER TABLE trials ADD COLUMN disease_zone_id INTEGER REFERENCES disease_zones(id)")
        missing_cols.append("disease_zone_id")

    if missing_cols:
        print(f"  已添加列: {', '.join(missing_cols)}")
    else:
        print("  所有列已存在，无需变更")

    # ---------------------------------------------------------------
    # 3. 设置所有旧数据 published=1
    # ---------------------------------------------------------------
    print("\n[3/5] 设置已有数据 published=1...")
    total = db.execute("SELECT COUNT(*) FROM trials").fetchone()[0]
    published = db.execute("SELECT COUNT(*) FROM trials WHERE published=1").fetchone()[0]
    pending = total - published
    if pending > 0:
        db.execute("UPDATE trials SET published=1, published_at=CURRENT_TIMESTAMP WHERE published=0 OR published IS NULL")
        print(f"  已公开 {pending} 条待审核数据 (总计 {total} 条)")
    else:
        print(f"  全部 {total} 条已公开，无需变更")

    # ---------------------------------------------------------------
    # 4. 插入初始管理员
    # ---------------------------------------------------------------
    print("\n[4/5] 插入初始管理员...")
    existing = db.execute("SELECT COUNT(*) FROM admin_users").fetchone()[0]
    if existing == 0:
        pwd_hash = "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9"
        db.execute("INSERT INTO admin_users(username, password_hash) VALUES(?, ?)",
                   ("admin", pwd_hash))
        print("  已创建管理员: admin / admin123")
    else:
        print(f"  管理员已存在 ({existing} 个)，跳过")

    # ---------------------------------------------------------------
    # 5. 迁移 crawl_sessions 旧数据
    # ---------------------------------------------------------------
    print("\n[5/5] 检查旧 crawl_sessions 表...")
    has_old = db.execute(
        "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='crawl_sessions'"
    ).fetchone()[0]
    if has_old:
        old_cnt = db.execute("SELECT COUNT(*) FROM crawl_sessions").fetchone()[0]
        if old_cnt > 0:
            db.execute("""
                INSERT OR IGNORE INTO crawl_logs(keyword, pages, found, new_items, start_time, end_time, status)
                SELECT keyword, pages_crawled, trials_found, trials_new, start_time, end_time, status
                FROM crawl_sessions
            """)
            print(f"  已迁移 {old_cnt} 条旧爬虫日志")
        else:
            print("  crawl_sessions 为空，无需迁移")

    db.execute("PRAGMA foreign_keys=ON")
    db.commit()
    db.close()
    print("\n==  迁移完成")


if __name__ == "__main__":
    if len(sys.argv) > 1:
        path = sys.argv[1]
    else:
        path = str(Path(__file__).resolve().parent.parent / "data" / "db" / "trials.db")
    migrate(path)
