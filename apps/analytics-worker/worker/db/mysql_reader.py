import os, pymysql

MYSQL_HOST = os.getenv('MYSQL_HOST', 'event-service-mysql')
MYSQL_PORT = int(os.getenv('MYSQL_PORT', '3306'))
MYSQL_USER = os.getenv('MYSQL_USER', 'root')
MYSQL_PASSWORD = os.getenv('MYSQL_PASSWORD', 'changeme-mysql-root')
MYSQL_DB = os.getenv('MYSQL_DB', 'events')

def get_mysql_connection():
    return pymysql.connect(host=MYSQL_HOST, port=MYSQL_PORT, user=MYSQL_USER,
        password=MYSQL_PASSWORD, database=MYSQL_DB, cursorclass=pymysql.cursors.DictCursor)

def get_events_by_org(org_id):
    conn = get_mysql_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT * FROM events WHERE org_id = %s ORDER BY start_time DESC", (org_id,))
            return cur.fetchall()
    finally:
        conn.close()

def get_revenue_by_event(event_id):
    conn = get_mysql_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("""SELECT t.name, t.price, COUNT(b.id) as sold, (t.price * COUNT(b.id)) as revenue
                FROM bookings b JOIN ticket_types t ON t.id = b.ticket_type_id
                WHERE b.event_id = %s AND b.status = 'confirmed'
                GROUP BY t.id, t.name, t.price""", (event_id,))
            return cur.fetchall()
    finally:
        conn.close()
