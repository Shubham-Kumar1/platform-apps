import os, signal, sys
import structlog
from celery import Celery

log = structlog.get_logger()
REDIS_URL = os.getenv('REDIS_URL', 'redis://redis:6379/0')

app = Celery('eventhub_analytics')
app.config_from_object({
    'broker_url': REDIS_URL,
    'result_backend': REDIS_URL,
    'task_serializer': 'json',
    'result_serializer': 'json',
    'accept_content': ['json'],
    'timezone': 'UTC',
    'task_acks_late': True,
    'worker_prefetch_multiplier': 1,
    'task_reject_on_worker_lost': True,
    'broker_connection_retry_on_startup': True,
})

app.conf.beat_schedule = {
    'generate-daily-reports': {
        'task': 'worker.tasks.reports.generate_daily_reports',
        'schedule': 86400.0,
    },
    'aggregate-event-stats': {
        'task': 'worker.tasks.aggregation.aggregate_all_events',
        'schedule': 300.0,
    },
    'cleanup-old-entries': {
        'task': 'worker.tasks.cleanup.cleanup_old_entries',
        'schedule': 3600.0,
    },
}

import worker.tasks.reports
import worker.tasks.aggregation
import worker.tasks.cleanup

def graceful_shutdown(signum, frame):
    log.info("received_shutdown_signal", signal=signum)
    sys.exit(0)

signal.signal(signal.SIGTERM, graceful_shutdown)
signal.signal(signal.SIGINT, graceful_shutdown)

if __name__ == '__main__':
    app.worker_main(['worker', '--loglevel=info',
        '--concurrency=' + os.getenv('WORKER_CONCURRENCY', '4')])
