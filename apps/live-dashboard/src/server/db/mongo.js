const mongoose = require('mongoose');
const MONGO_URI = process.env.MONGO_URI || 'mongodb://mongodb:27017/live_dashboard';

async function connectMongo() {
  await mongoose.connect(MONGO_URI, { maxPoolSize: 10, serverSelectionTimeoutMS: 5000 });
}
function getDb() { return mongoose.connection.db; }

const checkinSchema = new mongoose.Schema({
  userId:   { type: String, required: true, index: true },
  eventId:  { type: String, required: true, index: true },
  userName: String,
  orgName:  String,
  location: { lat: Number, lng: Number, name: String },
  timestamp: { type: Date, default: Date.now },
});
checkinSchema.index({ eventId: 1, timestamp: -1 });

const feedEntrySchema = new mongoose.Schema({
  eventId:   { type: String, required: true, index: true },
  type:      { type: String, enum: ['checkin', 'milestone', 'announcement', 'alert'] },
  message:   String,
  metadata:  mongoose.Schema.Types.Mixed,
  timestamp: { type: Date, default: Date.now },
});

const Checkin = mongoose.model('Checkin', checkinSchema);
const FeedEntry = mongoose.model('FeedEntry', feedEntrySchema);
module.exports = { connectMongo, getDb, Checkin, FeedEntry };
