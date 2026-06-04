/**
 * @neuralops/pulumi — programmatic tenant and collector provisioning (GAP-INT-001).
 */
import * as pulumi from '@pulumi/pulumi';

export interface TenantArgs {
  name: string;
  region?: string;
}

export class Tenant extends pulumi.ComponentResource {
  public readonly tenantId: pulumi.Output<string>;

  constructor(name: string, args: TenantArgs, opts?: pulumi.ComponentResourceOptions) {
    super('neuralops:Tenant', name, {}, opts);
    this.tenantId = pulumi.output(args.name);
    this.registerOutputs({ tenantId: this.tenantId });
  }
}

export interface CollectorAgentArgs {
  tenantId: pulumi.Input<string>;
  cluster: string;
  version?: string;
}

export class CollectorAgent extends pulumi.ComponentResource {
  constructor(name: string, args: CollectorAgentArgs, opts?: pulumi.ComponentResourceOptions) {
    super('neuralops:CollectorAgent', name, {}, opts);
    this.registerOutputs({
      cluster: args.cluster,
      version: args.version ?? '1.0.0',
    });
  }
}
