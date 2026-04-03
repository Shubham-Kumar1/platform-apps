import os, psycopg2
from psycopg2.extras import RealDictCursor

CENTRAL_DB_URL = os.getenv('CENTRAL_DB_URL',
    'postgresql://ro_tenant_alpha:password@pgbouncer.platform-services:5432/eventhub_central')

def get_central_connection():
    return psycopg2.connect(CENTRAL_DB_URL, cursor_factory=RealDictCursor)

def get_all_organizations():
    conn = get_central_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT id, name, slug, plan_tier FROM organizations")
            return cur.fetchall()
    finally:
        conn.close()

def get_org_members(org_id):
    conn = get_central_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("""SELECT u.id, u.email, u.name, m.role FROM users u
                JOIN org_memberships m ON m.user_id = u.id
                WHERE m.org_id = %s AND u.is_active = true""", (org_id,))
            return cur.fetchall()
    finally:
        conn.close()
