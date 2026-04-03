# Tenant Onboarding

## Onboard
```bash
cp values/tenant-alpha.yaml values/tenant-newtenant.yaml
make onboard-tenant TENANT=newtenant
make deploy-apps TENANT=newtenant
```

## Verify
```bash
kubectl get all -n tenant-newtenant
kubectl get resourcequota -n tenant-newtenant
```

## Offboard
```bash
helm uninstall event-service live-dashboard analytics-worker -n tenant-newtenant
kubectl delete namespace tenant-newtenant
```
