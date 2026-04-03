const express = require('express');
const router = express.Router();
const { Checkin, FeedEntry } = require('../db/mongo');
const { getUserWithOrg } = require('../db/central');

router.post('/', async (req, res) => {
  try {
    const { userId, eventId, location } = req.body;
    const user = await getUserWithOrg(userId);
    if (!user) return res.status(404).json({ error: 'User not found' });

    const checkin = await Checkin.create({
      userId, eventId, userName: user.name, orgName: user.org_name, location,
    });
    const feedEntry = await FeedEntry.create({
      eventId, type: 'checkin',
      message: `${user.name} from ${user.org_name} just checked in!`,
      metadata: { userId, userName: user.name, orgName: user.org_name },
    });

    const io = req.app.get('io');
    io.to(`event:${eventId}`).emit('new-checkin', { checkin, feedEntry });

    const count = await Checkin.countDocuments({ eventId });
    if (count % 100 === 0) {
      const milestone = await FeedEntry.create({
        eventId, type: 'milestone', message: `${count} attendees checked in!`,
      });
      io.to(`event:${eventId}`).emit('milestone', milestone);
    }
    res.status(201).json(checkin);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/:eventId', async (req, res) => {
  const checkins = await Checkin.find({ eventId: req.params.eventId }).sort({ timestamp: -1 }).limit(100);
  res.json(checkins);
});

router.get('/:eventId/count', async (req, res) => {
  const count = await Checkin.countDocuments({ eventId: req.params.eventId });
  res.json({ eventId: req.params.eventId, count });
});

module.exports = router;
