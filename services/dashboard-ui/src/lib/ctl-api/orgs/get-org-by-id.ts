import { api } from "@/lib/api";
import type { TOrg } from "@/types";

export const getOrgById = ({ orgId }: { orgId: string }) =>
  api<TOrg>({
    path: `orgs/current`,
    orgId,
  });
