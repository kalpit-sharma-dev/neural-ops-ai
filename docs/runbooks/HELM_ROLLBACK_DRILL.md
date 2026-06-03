# Helm Rollback Drill (Gate D)

**Target:** Restore previous release in **&lt; 15 minutes** wall-clock during approved change window.

## Prerequisites

- `kubectl` + `helm` authenticated to customer cluster
- Release name `neuralops` in namespace `neuralops`
- Previous revision known good (from last prod deploy)

## Dry run

```bash
./scripts/helm-rollback-drill.sh --dry-run
```

## Live drill steps

1. Note current revision: `helm history neuralops -n neuralops`
2. Start timer
3. Rollback one revision:

```bash
helm rollback neuralops <REVISION> -n neuralops --wait --timeout 10m
```

4. Verify:

```bash
kubectl rollout status deployment/gateway -n neuralops
curl -fsS https://$API_HOST/health
curl -fsS https://$API_HOST/ready
```

5. Stop timer — record elapsed seconds in Gate D evidence

## Failure handling

| Issue | Action |
|-------|--------|
| DB migration incompatible | Restore Postgres PITR per [BACKUP_RESTORE.md](./BACKUP_RESTORE.md); **do not** force rollback only |
| Pods ImagePullBackOff | Check registry credentials Secret |
| SSO broken after rollback | Re-sync OIDC client redirect URLs |

## Evidence template

```
Drill date: 
Operator: 
From revision: 
To revision: 
Elapsed (sec): 
Health OK: Y/N
Customer approver: 
```
