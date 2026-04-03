# Runbook: Credential Rotation

## 1. Generate new password
```bash
NEW_PASS=$(openssl rand -base64 24)
```

## 2. Update PostgreSQL
```bash
kubectl exec -it central-pg-0 -n platform-services -- \
  psql -U postgres -c "ALTER USER platform_admin PASSWORD '$NEW_PASS';"
```

## 3. Restart PgBouncer
```bash
kubectl rollout restart deployment/pgbouncer -n platform-services
```

## 4. Update tenant secrets and restart apps
```bash
for TENANT in alpha beta gamma; do
  kubectl rollout restart deployment -n tenant-$TENANT
done
```
