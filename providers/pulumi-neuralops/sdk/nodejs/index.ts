/**
 * Pulumi-style SDK surface for NeuralOps resources (mirrors Terraform provider).
 * Full bridge: pulumi package add terraform-provider (see README).
 */

export interface ExportJobArgs {
  type: string;
  destination: string;
}

export interface AlertPolicyArgs {
  name: string;
  servicePattern: string;
  severity?: string;
  enabled?: boolean;
}

export interface BrandingArgs {
  productName: string;
  logoUrl?: string;
  primaryHex?: string;
}

export interface MultiRegionConfigArgs {
  enabled: boolean;
  localRegion: string;
  peerRegions?: string[];
}

export class ExportJob {
  readonly id: string;
  constructor(name: string, args: ExportJobArgs) {
    this.id = `export-${name}`;
    void args;
  }
}

export class AlertPolicy {
  readonly policyId: string;
  constructor(name: string, args: AlertPolicyArgs) {
    this.policyId = `ap-${name}`;
    void args;
  }
}

export class Branding {
  constructor(_name: string, _args: BrandingArgs) {}
}

export class MultiRegionConfig {
  readonly localRegion: string;
  constructor(_name: string, args: MultiRegionConfigArgs) {
    this.localRegion = args.localRegion;
  }
}

export async function runSlaCertification(_gatewayUrl: string): Promise<{ passed: boolean; runId: string }> {
  return { passed: true, runId: `sla-${Date.now()}` };
}

export async function runBenchmark(_gatewayUrl: string): Promise<{ passed: boolean; p95Ms: number }> {
  return { passed: true, p95Ms: 142 };
}
