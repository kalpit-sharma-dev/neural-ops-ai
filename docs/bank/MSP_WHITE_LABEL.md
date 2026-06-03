# MSP & White-Label (Wave 6.3)

**ID:** ADM-03

## APIs

| Endpoint | Purpose |
|----------|---------|
| `GET/PUT /api/v1/admin/branding` | Logo, colors, product name, support email |
| `GET /api/v1/admin/msp/tenants` | List child tenants |
| `POST /api/v1/admin/msp/tenants` | Register child tenant |

## Branding example

```json
{
  "productName": "Acme Observe",
  "logoUrl": "https://cdn.acme.com/logo.svg",
  "primaryColor": "#0f766e",
  "accentColor": "#14b8a6",
  "supportEmail": "observe-support@acme.com",
  "customDomain": "observe.acme.com"
}
```

## Create child tenant

```bash
curl -s -X POST "$GATEWAY/api/v1/admin/msp/tenants" \
  -H "Authorization: Bearer $MSP_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Globex Bank POC","slug":"globex","plan":"enterprise","region":"eu-west-1"}'
```

## UI

**Settings → Enterprise governance** — tabs Branding, MSP tenants.

## Helm / ingress

Set `ingress.hosts.app` to MSP vanity domain; TLS cert per tenant region.

## Bank note

Tier-1 banks typically **do not** use multi-tenant MSP mode on shared SaaS — use **dedicated** or **self-hosted** single-tenant; MSP APIs for reseller partners only.
