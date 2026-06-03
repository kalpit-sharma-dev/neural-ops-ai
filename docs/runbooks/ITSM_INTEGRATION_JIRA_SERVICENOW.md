# ITSM Integration — Jira & ServiceNow (Wave 2.3)

**ID:** INT-01

## Overview

| System | Connect | Alert → ticket |
|--------|---------|----------------|
| Jira | OAuth (`/api/v1/integrations/jira/oauth/start`) | Workflow step `jira` / `jira_ticket` |
| ServiceNow | OAuth (instance URL in config) | Workflow webhook step (see below) |
| Slack / PagerDuty | OAuth or webhook | Workflow `slack`, `pagerduty` |

## Prerequisites

1. Integration row created in UI **Settings → Integrations** or API.
2. For OAuth: set `oauthClientId` (and secret in K8s Secret / External Secrets).
3. Gateway reachable from IdP redirect URL: `https://<customer-domain>/api/v1/integrations/oauth/callback`.

## Jira — configure

```json
{
  "integrationKey": "jira",
  "config": {
    "oauthClientId": "<atlassian-client-id>",
    "cloudId": "<atlassian-cloud-id>",
    "projectKey": "OPS",
    "issueType": "Incident"
  }
}
```

Connect:

```bash
curl -s -o /dev/null -w "%{http_code}" \
  "$GATEWAY/api/v1/integrations/jira/oauth/start" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT"
# Expect 302 to Atlassian; complete in browser
```

## ServiceNow — configure

```json
{
  "integrationKey": "servicenow",
  "config": {
    "instanceUrl": "https://acme.service-now.com",
    "oauthClientId": "<sn-client-id>",
    "assignmentGroup": "Observability-OnCall",
    "table": "incident"
  }
}
```

OAuth start: `GET /api/v1/integrations/servicenow/oauth/start`

For **event-driven** incidents without OAuth, use a workflow HTTP step targeting:

`POST https://<instance>/api/now/table/incident` with Basic or OAuth token stored in integration secret.

## Alert routing workflow

Example workflow trigger: `alert.fired`

```json
{
  "name": "P1 → Jira + ServiceNow",
  "triggerName": "alert.fired",
  "enabled": true,
  "steps": [
    { "id": "1", "type": "jira_ticket", "label": "Create Jira incident" },
    { "id": "2", "type": "servicenow", "label": "Create ServiceNow incident" }
  ]
}
```

Executor supports `jira`, `slack`, `pagerduty`, and `servicenow` natively when integration config includes `baseUrl`, `username`, and `password` (or `apiToken`).

## Verification script

```bash
chmod +x scripts/verify-itsm-integration.sh
GATEWAY=https://observe.customer.internal \
TOKEN=$POC_TOKEN \
TENANT=poc-bank \
./scripts/verify-itsm-integration.sh
```

## POC test plan

- [ ] Fire test alert on `ledger-service` severity P2
- [ ] Workflow creates Jira issue with trace link in description
- [ ] ServiceNow incident number returned or visible in workflow run log
- [ ] Disconnect OAuth revokes tokens (`POST .../integrations/jira/disconnect`)

## Troubleshooting

| Symptom | Check |
|---------|--------|
| 400 oauth not supported | Integration key must be `jira` or `servicenow` |
| 400 set oauthClientId first | ConfigMap / integration config |
| 302 loop | Redirect URI must match Atlassian/SN app registration |
