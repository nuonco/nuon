import { TNode } from "../types";

export type TApp = {
  name: string;
} & TNode;

export type TComponent = {
  config: any;
  name: string;
  type: "PUBLIC_IMAGE" | "HELM" | "TERRAFORM" | "GITHUB_REPO";
  vcsConfig?: Record<string, unknown>;
} & TNode;

export type TDeployment = {
  commitAuthor: string;
  commitHash: string;
} & TNode;

export type TRepo = {
  defaultBranch: string;
  fullName: string;
  name: string;
  owner: string;
  url: string;
};

export type TInstall = {
  name: string;
  settings: any;
} & TNode;

export type TOrg = {
  name: string;
} & TNode;

export type TSecret = {
  key: string;
  value: string;
};
