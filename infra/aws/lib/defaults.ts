import { quota_code, service_code } from "./enums";

export const allowedRegions = [
  "us-east-1",
  "us-east-2",
  "us-west-1",
  "us-west-2",
];

export const rootDomain = "nuon.co";

export const defaultRegion = "us-east-2";

type quota = {
  quotaCode: string;
  serviceCode: service_code;
  value: number;
};

type account = {
  // NOTE(jdt): this should be used judiciously and sparringly
  // we want to restrict e.g. prod resources in the stage account
  // abusing the additional tags would circumvent that
  additionalEnvTags?: string[];
  name: string;
  quotas?: quota[];
};

type unit = {
  accounts: account[];
  defaultQuotas?: quota[];
  disableScp?: boolean;
  name: string;
};

type orgStructure = unit[];

export const desiredOrgStructure: orgStructure = [
  {
    accounts: [],
    disableScp: true,
    name: "deleted",
  },
  {
    accounts: [
      { additionalEnvTags: ["stage", "prod"], name: "external" },
    ],
    name: "meta",
  },
  {
    accounts: [
      { name: "demo" },
      { name: "demo-govcloud" },
      { name: "public" },
      { name: "canary" },
    ],
    disableScp: true,
    name: "nuon-testing",
  },
  {
    accounts: [
      { name: "prod" },
      { name: "stage" },
      { additionalEnvTags: ["stage"], name: "infra-shared-stage" },
      { additionalEnvTags: ["prod"], name: "infra-shared-prod" },
      { additionalEnvTags: ["stage"], name: "govcloud-stage" },
      { additionalEnvTags: ["prod"], name: "govcloud-prod" },
      {
        additionalEnvTags: ["stage"],
        name: "orgs-stage",
        // NOTE(jdt): these somehow got set really high, have to set them here to match...
        quotas: [
          {
            quotaCode:
              quota_code.RUNNING_ON_DEMAND_STANDARD_A_C_D_H_I_M_R_T_Z_INSTANCES,
            serviceCode: service_code.EC2,
            value: 512,
          },
          {
            quotaCode:
              quota_code.ALL_STANDARD_A_C_D_H_I_M_R_T_Z_SPOT_INSTANCE_REQUESTS,
            serviceCode: service_code.EC2,
            value: 512,
          },
        ],
      },
      { additionalEnvTags: ["prod"], name: "orgs-prod" },
    ],
    defaultQuotas: [
      {
        quotaCode:
          quota_code.RUNNING_ON_DEMAND_STANDARD_A_C_D_H_I_M_R_T_Z_INSTANCES,
        serviceCode: service_code.EC2,
        value: 512,
      },
      {
        quotaCode:
          quota_code.ALL_STANDARD_A_C_D_H_I_M_R_T_Z_SPOT_INSTANCE_REQUESTS,
        serviceCode: service_code.EC2,
        value: 512,
      },
    ],
    name: "workloads",
  },
  {
    accounts: [
      { name: "sandbox-jm" },
      { name: "sandbox-ja" },
    ],
    disableScp: true,
    name: "engineers",
  },
];

export function getAccountQuotas(): Record<string, quota[]> {
  const out: Record<string, quota[]> = {};

  desiredOrgStructure.forEach((unit) => {
    unit.accounts.forEach((acct) => {
      const effectiveQuotas: { [name: string]: quota } = {};
      if (acct.quotas != undefined) {
        acct.quotas.forEach((q) => {
          effectiveQuotas[q.quotaCode] = q;
        });
      }
      if (unit.defaultQuotas != undefined) {
        unit.defaultQuotas.forEach((q) => {
          if (effectiveQuotas[q.quotaCode] == undefined) {
            effectiveQuotas[q.quotaCode] = q;
          }
        });
      }
      out[acct.name] = Object.entries(effectiveQuotas).map(([, q]) => q);
    });
  });

  return out;
}
