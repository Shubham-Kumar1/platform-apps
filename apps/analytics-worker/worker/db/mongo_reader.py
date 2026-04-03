import os
from pymongo import MongoClient

MONGO_URI = os.getenv('MONGO_URI', 'mongodb://live-dashboard-mongodb:27017/live_dashboard')
_client = None

def get_mongo_db():
    global _client
    if _client is None:
        _client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=5000)
    return _client.get_default_database()

def get_checkin_count_by_event(event_id):
    return get_mongo_db().checkins.count_documents({'eventId': event_id})

def get_hourly_checkin_distribution(event_id):
    return list(get_mongo_db().checkins.aggregate([
        {'$match': {'eventId': event_id}},
        {'$group': {'_id': {'$hour': '$timestamp'}, 'count': {'$sum': 1}}},
        {'$sort': {'_id': 1}}
    ]))
