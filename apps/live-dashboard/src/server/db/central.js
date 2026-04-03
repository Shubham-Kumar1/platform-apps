const { Pool } = require('pg');
const pool = new Pool({
  connectionString: process.env.CENTRAL_DB_URL ||
    'postgresql://ro_tenant_alpha:password@pgbouncer.platform-services:5432/eventhub_central',
  max: 5,
  idleTimeoutMillis: 30000,
});
function getCentralPool() { return pool; }

async function getUserWithOrg(userId) {
  const result = await pool.query(
    `SELECT u.id, u.email, u.name, o.name as org_name, m.role
     FROM users u
     JOIN org_memberships m ON m.user_id = u.id
     JOIN organizations o ON o.id = m.org_id
     WHERE u.id = $1 AND u.is_active = true LIMIT 1`, [userId]);
  return result.rows[0] || null;
}
module.exports = { getCentralPool, getUserWithOrg };
