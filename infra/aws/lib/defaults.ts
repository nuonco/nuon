import { quota_code, service_code } from "./enums";

const STAGE_REDUCED_EC2_QUOTAS = [
  {
    quotaCode:
      quota_code.RUNNING_ON_DEMAND_STANDARD_A_C_D_H_I_M_R_T_Z_INSTANCES,
    serviceCode: service_code.EC2,
    value: 512,
  },
  {
    quotaCode: quota_code.ALL_STANDARD_A_C_D_H_I_M_R_T_Z_SPOT_INSTANCE_REQUESTS,
    serviceCode: service_code.EC2,
    value: 512,
  },
];

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
    name: "nuon-public",
  },

  {
    accounts: [
      { name: "demo" },
      { name: "demo-govcloud" },
      { name: "public" },
      { name: "test" },
      { name: "infra-tests-00" },
      { name: "infra-tests-01" },
      { name: "infra-tests-02" },
      { name: "infra-tests-03" },
      { name: "infra-tests-04" },
    ],
    disableScp: true,
    name: "nuon-testing",
  },
  {
    accounts: [
      { name: "prod" },
      { name: "stage" },
      { name: "infra-shared-ci" },
      { additionalEnvTags: ["stage"], name: "infra-shared-stage" },
      { additionalEnvTags: ["prod"], name: "infra-shared-prod" },
      { additionalEnvTags: ["stage"], name: "govcloud-stage" },
      { additionalEnvTags: ["prod"], name: "govcloud-prod" },
      {
        additionalEnvTags: ["stage"],
        name: "orgs-stage",
        // NOTE(jdt): these somehow got set really high, have to set them here to match...
        quotas: STAGE_REDUCED_EC2_QUOTAS,
      },
      { additionalEnvTags: ["prod"], name: "orgs-prod" },
      // NOTE(fd): accounts for runners
      { additionalEnvTags: ["stage"], name: "runners-stage" },
      { additionalEnvTags: ["prod"], name: "runners-prod" },
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
      { name: "sandbox-rb" },
      { name: "sandbox-fd" },
      { name: "sandbox-nh" },
      { name: "sandbox-ht" },
      { name: "sandbox-cb" },
      { name: "sandbox-sk" },
      { name: "sandbox-mm" },
      { name: "sandbox-pk" },
      { name: "sandbox-ey" },
      { name: "sandbox-se" },
      { name: "sandbox-byoc" },
      { name: "sandbox-am" },

      // NOTE(jm): these accounts can be deprecated
      { name: "sandbox-retool" },
      { name: "sandbox-stardog" },
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
