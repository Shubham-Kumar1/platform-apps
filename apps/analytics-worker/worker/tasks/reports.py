import json, os, structlog
from datetime import datetime
from worker.main import app
from worker.db.central import get_all_organizations, get_org_members
from worker.db.mysql_reader import get_events_by_org, get_revenue_by_event
from worker.db.mongo_reader import get_checkin_count_by_event

log = structlog.get_logger()

@app.task(bind=True, name='worker.tasks.reports.generate_daily_reports')
def generate_daily_reports(self):
    orgs = get_all_organizations()
    for org in orgs:
        generate_org_report.delay(org['id'], org['name'])
    log.info("daily_reports_queued", count=len(orgs))

@app.task(bind=True, name='worker.tasks.reports.generate_org_report', max_retries=3)
def generate_org_report(self, org_id, org_name):
    try:
        events = get_events_by_org(org_id)
        members = get_org_members(org_id)
        report = {
            'org_id': org_id, 'org_name': org_name,
            'generated_at': datetime.utcnow().isoformat(),
            'summary': {'total_events': len(events), 'total_members': len(members)},
            'events': [],
        }
        for event in events[:20]:
            eid = event['id']
            checkins = get_checkin_count_by_event(eid)
            revenue_data = get_revenue_by_event(eid)
            total_rev = sum(float(r.get('revenue', 0)) for r in revenue_data)
            report['events'].append({
                'id': eid, 'title': event['title'], 'checkins': checkins, 'revenue': total_rev,
            })
        from redis import Redis
        redis = Redis.from_url(os.getenv('REDIS_URL', 'redis://redis:6379/0'))
        redis.setex(f'report:{org_id}:latest', 86400, json.dumps(report, default=str))
        log.info("report_generated", org=org_name)
    except Exception as e:
        log.error("report_failed", org=org_name, error=str(e))
        raise self.retry(exc=e)
