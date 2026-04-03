import json, os, structlog
from redis import Redis
from worker.main import app
from worker.db.mysql_reader import get_mysql_connection
from worker.db.mongo_reader import get_mongo_db

log = structlog.get_logger()

@app.task(name='worker.tasks.aggregation.aggregate_all_events')
def aggregate_all_events():
    redis = Redis.from_url(os.getenv('REDIS_URL', 'redis://redis:6379/0'))
    conn = get_mysql_connection()
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT id, org_id, title FROM events WHERE status = 'published'")
            events = cur.fetchall()
    finally:
        conn.close()

    mongo_db = get_mongo_db()
    for event in events:
        eid = event['id']
        checkins = mongo_db.checkins.count_documents({'eventId': eid})
        redis.setex(f'stats:event:{eid}', 600, json.dumps({
            'event_id': eid, 'title': event['title'], 'checkins': checkins,
        }))
    log.info("aggregation_complete", count=len(events))
