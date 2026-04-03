import structlog
from datetime import datetime, timedelta
from worker.main import app
from worker.db.mongo_reader import get_mongo_db

log = structlog.get_logger()

@app.task(name='worker.tasks.cleanup.cleanup_old_entries')
def cleanup_old_entries():
    db = get_mongo_db()
    cutoff = datetime.utcnow() - timedelta(days=30)
    result = db.feedentries.delete_many({'timestamp': {'$lt': cutoff}})
    log.info("cleanup_complete", deleted=result.deleted_count)
