import { App, RemoteBackend } from "cdktf";
import { Accounts, Org, SSO } from "./lib";

const defaultTags = [
  {
    tags: { environment: "management", terraform: "infra-aws" },
  },
];

const app = new App();
const org = new Org(app, "org", { defaultTags });
new RemoteBackend(org, {
  hostname: "app.terraform.io",
  organization: "nuonco",
  workspaces: {
    name: "aws-org",
  },
});

const sso = new SSO(app, "sso", {
  accounts: org.accounts,
  defaultTags,
});
new RemoteBackend(sso, {
  hostname: "app.terraform.io",
  organization: "nuonco",
  workspaces: {
    name: "aws-sso",
  },
});

const accounts = new Accounts(app, "accounts", {
  accounts: org.accounts,
  defaultTags,
});
new RemoteBackend(accounts, {
  hostname: "app.terraform.io",
  organization: "nuonco",
  workspaces: {
    name: "aws-accounts",
  },
});

app.synth();
