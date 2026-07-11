import sqlite3
import sys
# Set encoding for stdout
sys.stdout.reconfigure(encoding='utf-8') if hasattr(sys.stdout, 'reconfigure') else None

db = sqlite3.connect(r'C:\Users\hanq\projects\drug_trials_tool\data\db\trials.db')
# Check current encoding
cur = db.execute("SELECT hex(name) FROM disease_zones")
for row in cur:
    print("Before:", row[0])

# Write UTF-8 bytes directly using hex
name_utf8 = 'E883B0E885BAE7998C'.lower()       # 胰腺癌 in UTF-8 hex
keyword_utf8 = 'E883B0E885BA'.lower()            # 胰腺 in UTF-8 hex
desc_utf8 = 'E883B0E885BAE7998CE79BB8E585B3E4B8B4E5BA8AE8AF95E9AA8C'.lower()  # 胰腺癌相关临床试验 in UTF-8 hex

db.execute("UPDATE disease_zones SET name=x'" + name_utf8 + "', keyword=x'" + keyword_utf8 + "', description=x'" + desc_utf8 + "' WHERE id=1")
db.commit()

cur = db.execute("SELECT hex(name), hex(keyword) FROM disease_zones")
for row in cur:
    print("After:", row[0], row[1])
db.close()
