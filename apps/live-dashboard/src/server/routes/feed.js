const express = require('express');
const router = express.Router();
const { FeedEntry } = require('../db/mongo');

router.get('/:eventId', async (req, res) => {
  const { type, limit = 50 } = req.query;
  const query = { eventId: req.params.eventId };
  if (type) query.type = type;
  const entries = await FeedEntry.find(query).sort({ timestamp: -1 }).limit(parseInt(limit));
  res.json(entries);
});

router.post('/:eventId/announcement', async (req, res) => {
  const entry = await FeedEntry.create({
    eventId: req.params.eventId, type: 'announcement', message: req.body.message,
  });
  req.app.get('io').to(`event:${req.params.eventId}`).emit('announcement', entry);
  res.status(201).json(entry);
});

module.exports = router;
