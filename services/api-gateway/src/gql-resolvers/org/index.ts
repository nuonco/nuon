import { apps } from "../app/apps";
import { deleteOrg } from "./delete-org";
import { org } from "./org";
import { orgStatus } from "./org-status";
import { orgs } from "./orgs";
import { upsertOrg } from "./upsert-org";

export const orgResolvers = {
  Mutation: {
    deleteOrg,
    upsertOrg,
  },
  Org: {
    apps: (parent, { options }, ctx) =>
      apps(undefined, { options, orgId: parent.id }, ctx),
  },
  Query: {
    org,
    orgs,
    orgStatus,
  },
};
