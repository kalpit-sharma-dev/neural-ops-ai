# SSO Setup for Banks (BANK-007)

**Version:** 1.0

---

## 1. OIDC (recommended: Okta / Azure AD)

### Azure AD

1. App registration → Web → Redirect URI: `https://[app-host]/auth/callback`
2. Create client secret → store in External Secret.
3. Configure gateway:

```yaml
OIDC_ENABLED: "true"
OIDC_ISSUER: "https://login.microsoftonline.com/{tenant}/v2.0"
OIDC_CLIENT_ID: "{app-id}"
OIDC_CLIENT_SECRET: "{secret}"
OIDC_REDIRECT_URL: "https://app.bank.example/auth/callback"
```

4. Map groups to roles via IdP claims (custom claim `roles` or group → role mapping in gateway).

### Okta

Same pattern with Okta issuer URL and authorization code flow.

---

## 2. SAML 2.0

1. Upload SP metadata from `GET /auth/saml/metadata` (gateway).
2. Configure IdP ACS URL: `https://api.bank.example/auth/saml/acs`
3. Set:

```yaml
SAML_ENABLED: "true"
SAML_METADATA_URL: "https://idp.bank.example/metadata"
SAML_ENTITY_ID: "neuralops-bank-prod"
```

Use `crewjam/saml` path in production (`SAML_METADATA_URL` set).

---

## 3. Verification

```bash
./scripts/verify-keycloak-oidc.sh   # contract test against IdP
./scripts/verify-saml-metadata.sh
./scripts/verify-production-auth.sh
```

Manual: login as bank test user → confirm role and tenant.

---

## 4. MFA

Enforce at IdP — NeuralOps relies on IdP session strength.

---

*Disable dev login: [PRODUCTION_AUTH_CHECKLIST.md](../bank/PRODUCTION_AUTH_CHECKLIST.md)*
