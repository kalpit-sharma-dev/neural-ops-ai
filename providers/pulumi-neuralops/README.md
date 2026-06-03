# Pulumi bridge for NeuralOps

This package documents the **Pulumi Terraform bridge** pattern for the NeuralOps Terraform provider (`terraform-provider-neuralops`).

## Prerequisites

- Go 1.25+
- Pulumi CLI 3.x
- Built Terraform provider binary on `PATH` or `dev_overrides` in `~/.terraformrc`

## Generate the bridge (one-time)

```bash
cd providers/pulumi-neuralops
pulumi package add terraform-provider ../terraform-provider-neuralops
```

This emits a Node.js/Go/Python SDK that wraps:

- `neuralops_export_job`
- `neuralops_alert_policy`
- `neuralops_branding`

## Example (TypeScript)

```typescript
import * as neuralops from '@pulumi/neuralops';

const policy = new neuralops.AlertPolicy('checkout', {
  name: 'checkout-latency',
  servicePattern: 'payment-*',
  severity: 'P1',
});
```

## CI

Set `TF_ACC=1` and `NEURALOPS_API_URL` when running Terraform acceptance tests; Pulumi examples use the same API surface.
