const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const helmet = require('helmet');
const cors = require('cors');
const morgan = require('morgan');
const path = require('path');
const { connectMongo, getDb } = require('./db/mongo');
const { getCentralPool } = require('./db/central');
const checkinRoutes = require('./routes/checkin');
const feedRoutes = require('./routes/feed');

const app = express();
const server = http.createServer(app);
const io = new Server(server, { cors: { origin: '*' } });

app.use(helmet({ contentSecurityPolicy: false }));
app.use(cors());
app.use(morgan('combined'));
app.use(express.json());
app.use(express.static(path.join(__dirname, '../../dist')));

app.get('/health', (req, res) => res.json({ status: 'alive' }));
app.get('/ready', async (req, res) => {
  try {
    await getDb().command({ ping: 1 });
    await getCentralPool().query('SELECT 1');
    res.json({ status: 'ready', mongo: true, centralDb: true });
  } catch (err) {
    res.status(503).json({ status: 'not ready', error: err.message });
  }
});

app.set('io', io);
app.use('/api/checkins', checkinRoutes);
app.use('/api/feed', feedRoutes);

io.on('connection', (socket) => {
  console.log(`Client connected: ${socket.id}`);
  socket.on('join-event', (eventId) => {
    socket.join(`event:${eventId}`);
  });
  socket.on('leave-event', (eventId) => socket.leave(`event:${eventId}`));
  socket.on('disconnect', () => console.log(`Client disconnected: ${socket.id}`));
});

const PORT = process.env.PORT || 3000;
async function start() {
  await connectMongo();
  console.log('Connected to MongoDB');
  await getCentralPool().query('SELECT 1');
  console.log('Connected to Central PostgreSQL');
  server.listen(PORT, () => console.log(`Live Dashboard running on port ${PORT}`));
}

process.on('SIGTERM', () => {
  server.close(() => process.exit(0));
  setTimeout(() => process.exit(1), 10000);
});

start().catch((err) => { console.error('Startup failed:', err); process.exit(1); });
