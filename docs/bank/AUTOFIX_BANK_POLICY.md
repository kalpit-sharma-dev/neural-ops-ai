# AutoFix — Bank Safety Policy (Wave 6.1)

**ID:** AI-02

## Default for banks

| Setting | Production value |
|---------|------------------|
| `AUTOFIX_ENABLED` | `false` until CISO sign-off |
| `AUTOFIX_REQUIRE_APPROVAL` | `true` (always) |
| `AUTOFIX_ALLOWED_ROLES` | `ADMIN,SRE` only |

Helm prod: [values-prod.yaml](../../infra/helm/neuralops/values-prod.yaml) sets `autofixEnabled: false`.

## API flow

1. `POST /api/v1/ai/autofix/plan` — returns steps + `requiresApproval`
2. Human reviews in AIOps UI or ticket
3. `POST /api/v1/ai/autofix/execute` with `"approved": true` — only if policy allows
4. `POST /api/v1/ai/autofix/rollback` — revert within change window

## Blocks

- AutoFix disabled → **403** on all autofix routes
- `approved=false` when `AUTOFIX_REQUIRE_APPROVAL=true` → **403**
- Non-Admin/SRE role → **403**

## POC demo script

Enable only in lab tenant:

```bash
export AUTOFIX_ENABLED=true
export AUTOFIX_REQUIRE_APPROVAL=true
```

Never enable for production payment paths without change advisory board approval.

## Related

- [INCIDENT_RCA_PLAYBOOK.md](../runbooks/INCIDENT_RCA_PLAYBOOK.md)
