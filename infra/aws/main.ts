import { App, RemoteBackend } from "cdktf";
import { Accounts, Audit, Org, SSO } from "./lib";

const defaultTags = [
  {
    tags: { environment: "management", terraform: "infra-aws" },
  },
];

const app = new App();
const org = new Org(app, "org", { defaultTags });
new RemoteBackend(org, {
  hostname: "app.terraform.io",
  organization: "launchpaddev",
  workspaces: {
    name: "aws",
  },
});

const sso = new SSO(app, "sso", {
  accounts: org.accounts,
  defaultTags,
});
new RemoteBackend(sso, {
  hostname: "app.terraform.io",
  organization: "launchpaddev",
  workspaces: {
    name: "aws-sso",
  },
});

const auditAccount = org.accounts.find((acct) => acct.name == "audit");
if (!auditAccount) {
  throw new Error("unable to find audit account");
}
const audit = new Audit(app, "audit", {
  account: auditAccount,
  defaultTags,
  org: org.org,
});
new RemoteBackend(audit, {
  hostname: "app.terraform.io",
  organization: "launchpaddev",
  workspaces: {
    name: "aws-audit",
  },
});

const accounts = new Accounts(app, "accounts", {
  accounts: org.accounts,
  defaultTags,
});
new RemoteBackend(accounts, {
  hostname: "app.terraform.io",
  organization: "launchpaddev",
  workspaces: {
    name: "aws-accounts",
  },
});

app.synth();
