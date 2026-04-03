# Runbook: Central DB Failover

## Check pod status
```bash
kubectl get pods -n platform-services -l app=central-pg
kubectl describe pod central-pg-0 -n platform-services
kubectl logs central-pg-0 -n platform-services --tail=100
```

## Force restart if needed
```bash
kubectl delete pod central-pg-0 -n platform-services
```

## Verify PgBouncer reconnected
```bash
kubectl exec -it deploy/pgbouncer -n platform-services -- pgbouncer -R
```
