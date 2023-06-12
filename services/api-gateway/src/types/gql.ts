/* eslint-disable */
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string | number; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** Represents an ISO 8601-encoded date and time string. For example, 3:50 pm on September 7, 2019 in the time zone of UTC (Coordinated Universal Time) is represented as '2019-09-07T15:50:00Z' */
  DateTime: { input: any; output: any; }
};

/** Represents the AWS Authentication configuration used to access the vendor's OCI image */
export type AwsAuthConfig = {
  __typename?: 'AWSAuthConfig';
  /** AWS Region */
  region: AwsRegion;
  /** AWS IAM Role */
  role: Scalars['String']['output'];
};

export type AwsAuthConfigInput = {
  region: AwsRegion;
  role: Scalars['String']['input'];
};

export enum AwsRegion {
  UsEast_1 = 'US_EAST_1',
  UsEast_2 = 'US_EAST_2',
  UsWest_1 = 'US_WEST_1',
  UsWest_2 = 'US_WEST_2'
}

/** Represents a settings for AWS cloud target */
export type AwsSettings = {
  __typename?: 'AWSSettings';
  region: AwsRegion;
  role: Scalars['String']['output'];
};

export type AwsSettingsInput = {
  region: AwsRegion;
  role: Scalars['String']['input'];
};

/** Represents a collection of general settings and information about a User's software */
export type App = Node & {
  __typename?: 'App';
  components: ComponentConnection;
  createdAt: Scalars['DateTime']['output'];
  deployments: DeploymentConnection;
  id: Scalars['ID']['output'];
  installs: InstallConnection;
  name: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};


/** Represents a collection of general settings and information about a User's software */
export type AppComponentsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};


/** Represents a collection of general settings and information about a User's software */
export type AppDeploymentsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};


/** Represents a collection of general settings and information about a User's software */
export type AppInstallsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};

/** An auto-generated type for paginating through multiple Apps */
export type AppConnection = Connection & {
  __typename?: 'AppConnection';
  /** A list of edges */
  edges?: Maybe<Array<AppEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

/** An auto-generated type which holds one App and a cursor during pagination */
export type AppEdge = {
  __typename?: 'AppEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of AppEdge */
  node: App;
};

export type AppInput = {
  githubInstallId?: InputMaybe<Scalars['ID']['input']>;
  id?: InputMaybe<Scalars['ID']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  orgId?: InputMaybe<Scalars['ID']['input']>;
};

/** Represents a basic Kubernetes deployment configurations */
export type BasicDeployConfig = {
  __typename?: 'BasicDeployConfig';
  /** The health check endpoint for the Kubernetes liveness probe */
  healthCheckPath: Scalars['String']['output'];
  /** How many instances of deployment to maintain */
  instanceCount: Scalars['Int']['output'];
  /** Container port */
  port: Scalars['Int']['output'];
};

export type BasicDeployConfigInput = {
  healthCheckPath: Scalars['String']['input'];
  instanceCount: Scalars['Int']['input'];
  port: Scalars['Int']['input'];
};

/** Represents information about a build */
export type Build = {
  __typename?: 'Build';
  componentId: Scalars['ID']['output'];
  createdAt: Scalars['DateTime']['output'];
  gitRef: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type BuildConfigInput = {
  dockerBuildConfig?: InputMaybe<DockerBuildInput>;
  externalImageConfig?: InputMaybe<ExternalImageInput>;
  noop?: InputMaybe<Scalars['Boolean']['input']>;
  terraformBuildConfig?: InputMaybe<TerraformBuildInput>;
};

export type BuildInput = {
  componentId: Scalars['ID']['input'];
  gitRef?: InputMaybe<Scalars['String']['input']>;
};

/** Represents a collection of general settings and information about a piece of a App */
export type Component = Node & {
  __typename?: 'Component';
  app?: Maybe<App>;
  config?: Maybe<ComponentConfig>;
  createdAt: Scalars['DateTime']['output'];
  deployments: DeploymentConnection;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};


/** Represents a collection of general settings and information about a piece of a App */
export type ComponentDeploymentsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};

/** Represents all the component build configurations */
export type ComponentBuildConfig = DockerBuildConfig | ExternalImageConfig | NoopConfig | TerraformBuildConfig;

/** Represents all configuration for the  component */
export type ComponentConfig = {
  __typename?: 'ComponentConfig';
  buildConfig?: Maybe<ComponentBuildConfig>;
  deployConfig?: Maybe<ComponentDeployConfig>;
};

export type ComponentConfigInput = {
  buildConfig?: InputMaybe<BuildConfigInput>;
  deployConfig?: InputMaybe<DeployConfigInput>;
};

/** An auto-generated type for paginating through multiple Components */
export type ComponentConnection = Connection & {
  __typename?: 'ComponentConnection';
  /** A list of edges */
  edges?: Maybe<Array<ComponentEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

/** Represents all the component deployment configurations */
export type ComponentDeployConfig = BasicDeployConfig | HelmRepoDeployConfig | NoopConfig | TerraformDeployConfig;

/** An auto-generated type which holds one Component and a cursor during pagination */
export type ComponentEdge = {
  __typename?: 'ComponentEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of ComponentEdge */
  node?: Maybe<Component>;
};

export type ComponentInput = {
  appId?: InputMaybe<Scalars['ID']['input']>;
  config?: InputMaybe<ComponentConfigInput>;
  id?: InputMaybe<Scalars['ID']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

/** Represents settings for a connected github repository */
export type ConnectedGithubConfig = {
  __typename?: 'ConnectedGithubConfig';
  branch?: Maybe<Scalars['String']['output']>;
  directory: Scalars['String']['output'];
  gitRef?: Maybe<Scalars['String']['output']>;
  repo: Scalars['String']['output'];
};

export type ConnectedGithubConfigInput = {
  branch: Scalars['String']['input'];
  directory: Scalars['String']['input'];
  gitRef?: InputMaybe<Scalars['String']['input']>;
  repo: Scalars['String']['input'];
};

export type Connection = {
  /** Information to aid in pagination of list */
  pageInfo: PageInfo;
  /** Total count of items */
  totalCount: Scalars['Int']['output'];
};

export type ConnectionOptions = {
  /** Returns the elements that come after the specified cursor */
  after?: InputMaybe<Scalars['String']['input']>;
  /** Returns the elements that come before the specified cursor */
  before?: InputMaybe<Scalars['String']['input']>;
  /** Returns first (n) elements */
  first?: InputMaybe<Scalars['Int']['input']>;
  /** Returns last (n) elements */
  last?: InputMaybe<Scalars['Int']['input']>;
};

/** Represents information about a deploy */
export type Deploy = {
  __typename?: 'Deploy';
  buildId: Scalars['ID']['output'];
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  installId: Scalars['ID']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type DeployConfigInput = {
  basicDeployConfig?: InputMaybe<BasicDeployConfigInput>;
  helmRepoDeployConfig?: InputMaybe<HelmRepoDeployConfigInput>;
  noop?: InputMaybe<Scalars['Boolean']['input']>;
  terraformDeployConfig?: InputMaybe<TerraformDeployConfigInput>;
};

export type DeployInput = {
  buildId: Scalars['ID']['input'];
  installId: Scalars['ID']['input'];
};

/** Represents a collection of general settings and information about a deployed piece of an App */
export type Deployment = Node & {
  __typename?: 'Deployment';
  commitAuthor?: Maybe<Scalars['String']['output']>;
  commitHash?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

/** An auto-generated type for paginating through multiple Deployments */
export type DeploymentConnection = Connection & {
  __typename?: 'DeploymentConnection';
  /** A list of edges */
  edges?: Maybe<Array<DeploymentEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

/** An auto-generated type which holds one Deployment and a cursor during pagination */
export type DeploymentEdge = {
  __typename?: 'DeploymentEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of DeploymentEdge */
  node?: Maybe<Deployment>;
};

/** Represents a Docker build configuration */
export type DockerBuildConfig = {
  __typename?: 'DockerBuildConfig';
  /** Docker build arguments */
  buildArgs?: Maybe<Array<KeyValuePair>>;
  /** Name of Dockerfile used to build image */
  dockerfile: Scalars['String']['output'];
  /** Environment variables */
  envVarsConfig?: Maybe<Array<KeyValuePair>>;
  /** Version control system configuration */
  vcsConfig?: Maybe<VcsConfig>;
};

export type DockerBuildInput = {
  buildArgs?: InputMaybe<Array<KeyValuePairInput>>;
  dockerfile: Scalars['String']['input'];
  envVarsConfig?: InputMaybe<Array<KeyValuePairInput>>;
  vcsConfig: VcsConfigInput;
};

/** Represents an external OCI image configuration */
export type ExternalImageConfig = {
  __typename?: 'ExternalImageConfig';
  authConfig?: Maybe<AwsAuthConfig>;
  /** URL where the OCI image is hosted */
  ociImageUrl: Scalars['String']['output'];
  /** OCI image tag */
  tag: Scalars['String']['output'];
};

export type ExternalImageInput = {
  authConfig?: InputMaybe<AwsAuthConfigInput>;
  ociImageUrl: Scalars['String']['input'];
  tag?: InputMaybe<Scalars['String']['input']>;
};

/** Represents a settings for GCP cloud target */
export type GcpSettings = {
  __typename?: 'GCPSettings';
  noop: Scalars['Boolean']['output'];
};

export type GcpSettingsInput = {
  bogus: Scalars['String']['input'];
};

/** Represents a public helm chart deployment configuration */
export type HelmRepoDeployConfig = {
  __typename?: 'HelmRepoDeployConfig';
  chartName: Scalars['String']['output'];
  chartRepo: Scalars['String']['output'];
  chartVersion: Scalars['String']['output'];
  imageRepoValuesKey?: Maybe<Scalars['String']['output']>;
  imageTagValuesKey?: Maybe<Scalars['String']['output']>;
};

export type HelmRepoDeployConfigInput = {
  chartName: Scalars['String']['input'];
  chartRepo: Scalars['String']['input'];
  chartVersion: Scalars['String']['input'];
  imageRepoValuesKey?: InputMaybe<Scalars['String']['input']>;
  imageTagValuesKey?: InputMaybe<Scalars['String']['input']>;
};

/** Represents a collection of general settings and information about a cloud target */
export type Install = Node & {
  __typename?: 'Install';
  components: ComponentConnection;
  createdAt: Scalars['DateTime']['output'];
  deployments: DeploymentConnection;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  settings: InstallSettings;
  updatedAt: Scalars['DateTime']['output'];
};


/** Represents a collection of general settings and information about a cloud target */
export type InstallComponentsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};


/** Represents a collection of general settings and information about a cloud target */
export type InstallDeploymentsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};

/** An auto-generated type for paginating through multiple Installs */
export type InstallConnection = Connection & {
  __typename?: 'InstallConnection';
  /** A list of edges */
  edges?: Maybe<Array<InstallEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

/** An auto-generated type which holds one Install and a cursor during pagination */
export type InstallEdge = {
  __typename?: 'InstallEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of InstallEdge */
  node: Install;
};

export type InstallInput = {
  appId?: InputMaybe<Scalars['ID']['input']>;
  awsSettings?: InputMaybe<AwsSettingsInput>;
  gcpSettings?: InputMaybe<GcpSettingsInput>;
  id?: InputMaybe<Scalars['ID']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

/** Represents cloud target settings for AWS or GCP */
export type InstallSettings = AwsSettings | GcpSettings;

/** Represents information about an instance */
export type Instance = {
  __typename?: 'Instance';
  buildId: Scalars['ID']['output'];
  componentId: Scalars['ID']['output'];
  createdAt: Scalars['DateTime']['output'];
  deployId: Scalars['ID']['output'];
  id: Scalars['ID']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

/** Represents a collection of general info about deployed piece of software on an Install */
export type InstanceStatus = {
  __typename?: 'InstanceStatus';
  hostname?: Maybe<Scalars['String']['output']>;
  status: Status;
};

/** Represents a key value pair */
export type KeyValuePair = {
  __typename?: 'KeyValuePair';
  key: Scalars['String']['output'];
  value: Scalars['String']['output'];
};

export type KeyValuePairInput = {
  key: Scalars['String']['input'];
  value: Scalars['String']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  cancelBuild: Scalars['Boolean']['output'];
  createDeployment?: Maybe<Deployment>;
  deleteApp: Scalars['Boolean']['output'];
  deleteComponent: Scalars['Boolean']['output'];
  deleteInstall: Scalars['Boolean']['output'];
  deleteOrg: Scalars['Boolean']['output'];
  deleteSecrets: Scalars['Boolean']['output'];
  echo: Scalars['String']['output'];
  startBuild: Build;
  startDeploy: Deploy;
  upsertApp?: Maybe<App>;
  upsertComponent?: Maybe<Component>;
  upsertInstall?: Maybe<Install>;
  upsertOrg?: Maybe<Org>;
  upsertSecrets: Scalars['Boolean']['output'];
};


export type MutationCancelBuildArgs = {
  id: Scalars['ID']['input'];
};


export type MutationCreateDeploymentArgs = {
  componentId: Scalars['ID']['input'];
};


export type MutationDeleteAppArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteComponentArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteInstallArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteOrgArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteSecretsArgs = {
  input?: InputMaybe<Array<SecretsIdsInput>>;
};


export type MutationEchoArgs = {
  word: Scalars['String']['input'];
};


export type MutationStartBuildArgs = {
  input: BuildInput;
};


export type MutationStartDeployArgs = {
  input: DeployInput;
};


export type MutationUpsertAppArgs = {
  input: AppInput;
};


export type MutationUpsertComponentArgs = {
  input: ComponentInput;
};


export type MutationUpsertInstallArgs = {
  input: InstallInput;
};


export type MutationUpsertOrgArgs = {
  input: OrgInput;
};


export type MutationUpsertSecretsArgs = {
  input?: InputMaybe<Array<SecretsInput>>;
};

export type Node = {
  /** The date and time (ISO 8601 format) when the node was created */
  createdAt: Scalars['DateTime']['output'];
  /** A globally-unique identifier */
  id: Scalars['ID']['output'];
  /** The date and time (ISO 8601 format) when the node was last updated */
  updatedAt: Scalars['DateTime']['output'];
};

/** Represents no operation configuration */
export type NoopConfig = {
  __typename?: 'NoopConfig';
  noop: Scalars['Boolean']['output'];
};

export type NoopSettings = {
  __typename?: 'NoopSettings';
  noop?: Maybe<Scalars['Boolean']['output']>;
};

/** Represents a collection of general settings and information about a Nuon tenant */
export type Org = Node & {
  __typename?: 'Org';
  apps: AppConnection;
  createdAt: Scalars['DateTime']['output'];
  githubInstallId?: Maybe<Scalars['ID']['output']>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};


/** Represents a collection of general settings and information about a Nuon tenant */
export type OrgAppsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};

/** An auto-generated type for paginating through multiple Orgs */
export type OrgConnection = Connection & {
  __typename?: 'OrgConnection';
  /** A list of edges */
  edges?: Maybe<Array<OrgEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

/** An auto-generated type which holds one Org and a cursor during pagination */
export type OrgEdge = {
  __typename?: 'OrgEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of OrgEdge */
  node: Org;
};

export type OrgInput = {
  githubInstallId?: InputMaybe<Scalars['ID']['input']>;
  id?: InputMaybe<Scalars['ID']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  ownerId?: InputMaybe<Scalars['ID']['input']>;
};

/** Returns information about pagination in a connection, in accordance with the Relay specification */
export type PageInfo = {
  __typename?: 'PageInfo';
  /** The cursor corresponding to the last node in edges */
  endCursor?: Maybe<Scalars['String']['output']>;
  /** Whether there are more pages to fetch following the current page */
  hasNextPage: Scalars['Boolean']['output'];
  /** Whether there are any pages prior to the current page */
  hasPreviousPage: Scalars['Boolean']['output'];
  /** The cursor corresponding to the first node in edges */
  startCursor?: Maybe<Scalars['String']['output']>;
};

/** Represents settings for a public git repository */
export type PublicGitConfig = {
  __typename?: 'PublicGitConfig';
  directory: Scalars['String']['output'];
  gitRef?: Maybe<Scalars['String']['output']>;
  repo: Scalars['String']['output'];
};

export type PublicGitConfigInput = {
  directory: Scalars['String']['input'];
  gitRef?: InputMaybe<Scalars['String']['input']>;
  repo: Scalars['String']['input'];
};

export type Query = {
  __typename?: 'Query';
  app?: Maybe<App>;
  apps: AppConnection;
  build?: Maybe<Build>;
  buildStatus: Status;
  builds?: Maybe<Array<Build>>;
  component?: Maybe<Component>;
  components: ComponentConnection;
  deploy: Deploy;
  deployment?: Maybe<Deployment>;
  deploymentStatus: Status;
  deployments: DeploymentConnection;
  install?: Maybe<Install>;
  installStatus: Status;
  installs: InstallConnection;
  instanceStatus: InstanceStatus;
  instances?: Maybe<Array<Instance>>;
  me?: Maybe<User>;
  org?: Maybe<Org>;
  orgStatus: Status;
  orgs: OrgConnection;
  ping?: Maybe<Scalars['String']['output']>;
  repos: RepoConnection;
  secrets?: Maybe<Array<Secret>>;
};


export type QueryAppArgs = {
  id: Scalars['ID']['input'];
};


export type QueryAppsArgs = {
  options?: InputMaybe<ConnectionOptions>;
  orgId: Scalars['ID']['input'];
};


export type QueryBuildArgs = {
  id: Scalars['ID']['input'];
};


export type QueryBuildStatusArgs = {
  appId: Scalars['ID']['input'];
  buildId: Scalars['ID']['input'];
  componentId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
};


export type QueryBuildsArgs = {
  componentId: Scalars['ID']['input'];
};


export type QueryComponentArgs = {
  id: Scalars['ID']['input'];
};


export type QueryComponentsArgs = {
  appId: Scalars['ID']['input'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryDeployArgs = {
  id: Scalars['ID']['input'];
};


export type QueryDeploymentArgs = {
  id: Scalars['ID']['input'];
};


export type QueryDeploymentStatusArgs = {
  appId: Scalars['ID']['input'];
  componentId: Scalars['ID']['input'];
  deploymentId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
};


export type QueryDeploymentsArgs = {
  appIds?: InputMaybe<Array<InputMaybe<Scalars['ID']['input']>>>;
  componentIds?: InputMaybe<Array<InputMaybe<Scalars['ID']['input']>>>;
  installIds?: InputMaybe<Array<InputMaybe<Scalars['ID']['input']>>>;
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryInstallArgs = {
  id: Scalars['ID']['input'];
};


export type QueryInstallStatusArgs = {
  appId: Scalars['ID']['input'];
  installId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
};


export type QueryInstallsArgs = {
  appId: Scalars['ID']['input'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryInstanceStatusArgs = {
  appId: Scalars['ID']['input'];
  componentId: Scalars['ID']['input'];
  deploymentId: Scalars['ID']['input'];
  installId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
};


export type QueryInstancesArgs = {
  installId: Scalars['ID']['input'];
};


export type QueryOrgArgs = {
  id: Scalars['ID']['input'];
};


export type QueryOrgStatusArgs = {
  id: Scalars['ID']['input'];
};


export type QueryOrgsArgs = {
  memberId: Scalars['ID']['input'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryReposArgs = {
  githubInstallId: Scalars['ID']['input'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QuerySecretsArgs = {
  appId: Scalars['ID']['input'];
  componentId: Scalars['ID']['input'];
  installId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
};

export type Repo = {
  __typename?: 'Repo';
  defaultBranch?: Maybe<Scalars['String']['output']>;
  fullName?: Maybe<Scalars['String']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  owner?: Maybe<Scalars['String']['output']>;
  url?: Maybe<Scalars['String']['output']>;
};

/** An auto-generated type for paginating through multiple Installs */
export type RepoConnection = Connection & {
  __typename?: 'RepoConnection';
  /** A list of edges */
  edges?: Maybe<Array<RepoEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

/** An auto-generated type which holds one Install and a cursor during pagination */
export type RepoEdge = {
  __typename?: 'RepoEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of RepoEdge */
  node: Repo;
};

/** Represents a secret key-value configuration */
export type Secret = {
  __typename?: 'Secret';
  id: Scalars['ID']['output'];
  key: Scalars['String']['output'];
  value: Scalars['String']['output'];
};

export type SecretInput = {
  id?: InputMaybe<Scalars['ID']['input']>;
  key: Scalars['String']['input'];
  value: Scalars['String']['input'];
};

export type SecretsIdsInput = {
  appId: Scalars['ID']['input'];
  componentId: Scalars['ID']['input'];
  installId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
  secretId: Scalars['ID']['input'];
};

export type SecretsInput = {
  appId: Scalars['ID']['input'];
  componentId: Scalars['ID']['input'];
  installId: Scalars['ID']['input'];
  orgId: Scalars['ID']['input'];
  secrets?: InputMaybe<Array<SecretInput>>;
};

export enum Status {
  Active = 'ACTIVE',
  Error = 'ERROR',
  Provisioning = 'PROVISIONING',
  Unknown = 'UNKNOWN',
  Unspecified = 'UNSPECIFIED'
}

/** Represents a Terraform module build configuration */
export type TerraformBuildConfig = {
  __typename?: 'TerraformBuildConfig';
  /** Environment variables for the build */
  envVarsConfig?: Maybe<Array<Maybe<KeyValuePair>>>;
  /** Version control system configuration */
  vcsConfig?: Maybe<VcsConfig>;
};

export type TerraformBuildInput = {
  envVarsConfig?: InputMaybe<Array<KeyValuePairInput>>;
  vcsConfig: VcsConfigInput;
};

/** Represents a terraform module deployment configuration */
export type TerraformDeployConfig = {
  __typename?: 'TerraformDeployConfig';
  terraformVersion?: Maybe<TerraformVersion>;
};

export type TerraformDeployConfigInput = {
  terraformVersion?: InputMaybe<TerraformVersion>;
};

export enum TerraformVersion {
  TerraformVersion_0_8_8 = 'TERRAFORM_VERSION_0_8_8',
  TerraformVersion_0_9_1 = 'TERRAFORM_VERSION_0_9_1',
  TerraformVersion_0_9_2 = 'TERRAFORM_VERSION_0_9_2',
  TerraformVersion_0_9_3 = 'TERRAFORM_VERSION_0_9_3',
  TerraformVersion_0_9_4 = 'TERRAFORM_VERSION_0_9_4',
  TerraformVersion_0_9_5 = 'TERRAFORM_VERSION_0_9_5',
  TerraformVersion_0_9_6 = 'TERRAFORM_VERSION_0_9_6',
  TerraformVersion_0_9_7 = 'TERRAFORM_VERSION_0_9_7',
  TerraformVersion_0_9_9 = 'TERRAFORM_VERSION_0_9_9',
  TerraformVersion_0_9_10 = 'TERRAFORM_VERSION_0_9_10',
  TerraformVersion_0_9_11 = 'TERRAFORM_VERSION_0_9_11',
  TerraformVersion_0_10_1 = 'TERRAFORM_VERSION_0_10_1',
  TerraformVersion_0_10_2 = 'TERRAFORM_VERSION_0_10_2',
  TerraformVersion_0_10_3 = 'TERRAFORM_VERSION_0_10_3',
  TerraformVersion_0_10_4 = 'TERRAFORM_VERSION_0_10_4',
  TerraformVersion_0_10_5 = 'TERRAFORM_VERSION_0_10_5',
  TerraformVersion_0_10_6 = 'TERRAFORM_VERSION_0_10_6',
  TerraformVersion_0_10_7 = 'TERRAFORM_VERSION_0_10_7',
  TerraformVersion_0_10_8 = 'TERRAFORM_VERSION_0_10_8',
  TerraformVersion_0_11_1 = 'TERRAFORM_VERSION_0_11_1',
  TerraformVersion_0_11_2 = 'TERRAFORM_VERSION_0_11_2',
  TerraformVersion_0_11_3 = 'TERRAFORM_VERSION_0_11_3',
  TerraformVersion_0_11_4 = 'TERRAFORM_VERSION_0_11_4',
  TerraformVersion_0_11_5 = 'TERRAFORM_VERSION_0_11_5',
  TerraformVersion_0_11_6 = 'TERRAFORM_VERSION_0_11_6',
  TerraformVersion_0_11_7 = 'TERRAFORM_VERSION_0_11_7',
  TerraformVersion_0_11_8 = 'TERRAFORM_VERSION_0_11_8',
  TerraformVersion_0_11_9 = 'TERRAFORM_VERSION_0_11_9',
  TerraformVersion_0_11_10 = 'TERRAFORM_VERSION_0_11_10',
  TerraformVersion_0_11_11 = 'TERRAFORM_VERSION_0_11_11',
  TerraformVersion_0_11_12 = 'TERRAFORM_VERSION_0_11_12',
  TerraformVersion_0_11_13 = 'TERRAFORM_VERSION_0_11_13',
  TerraformVersion_0_11_14 = 'TERRAFORM_VERSION_0_11_14',
  TerraformVersion_0_11_15 = 'TERRAFORM_VERSION_0_11_15',
  TerraformVersion_0_12_1 = 'TERRAFORM_VERSION_0_12_1',
  TerraformVersion_0_12_2 = 'TERRAFORM_VERSION_0_12_2',
  TerraformVersion_0_12_3 = 'TERRAFORM_VERSION_0_12_3',
  TerraformVersion_0_12_4 = 'TERRAFORM_VERSION_0_12_4',
  TerraformVersion_0_12_5 = 'TERRAFORM_VERSION_0_12_5',
  TerraformVersion_0_12_6 = 'TERRAFORM_VERSION_0_12_6',
  TerraformVersion_0_12_7 = 'TERRAFORM_VERSION_0_12_7',
  TerraformVersion_0_12_8 = 'TERRAFORM_VERSION_0_12_8',
  TerraformVersion_0_12_9 = 'TERRAFORM_VERSION_0_12_9',
  TerraformVersion_0_12_10 = 'TERRAFORM_VERSION_0_12_10',
  TerraformVersion_0_12_11 = 'TERRAFORM_VERSION_0_12_11',
  TerraformVersion_0_12_12 = 'TERRAFORM_VERSION_0_12_12',
  TerraformVersion_0_12_13 = 'TERRAFORM_VERSION_0_12_13',
  TerraformVersion_0_12_14 = 'TERRAFORM_VERSION_0_12_14',
  TerraformVersion_0_12_15 = 'TERRAFORM_VERSION_0_12_15',
  TerraformVersion_0_12_16 = 'TERRAFORM_VERSION_0_12_16',
  TerraformVersion_0_12_17 = 'TERRAFORM_VERSION_0_12_17',
  TerraformVersion_0_12_18 = 'TERRAFORM_VERSION_0_12_18',
  TerraformVersion_0_12_19 = 'TERRAFORM_VERSION_0_12_19',
  TerraformVersion_0_12_20 = 'TERRAFORM_VERSION_0_12_20',
  TerraformVersion_0_12_21 = 'TERRAFORM_VERSION_0_12_21',
  TerraformVersion_0_12_22 = 'TERRAFORM_VERSION_0_12_22',
  TerraformVersion_0_12_23 = 'TERRAFORM_VERSION_0_12_23',
  TerraformVersion_0_12_24 = 'TERRAFORM_VERSION_0_12_24',
  TerraformVersion_0_12_25 = 'TERRAFORM_VERSION_0_12_25',
  TerraformVersion_0_12_26 = 'TERRAFORM_VERSION_0_12_26',
  TerraformVersion_0_12_27 = 'TERRAFORM_VERSION_0_12_27',
  TerraformVersion_0_12_28 = 'TERRAFORM_VERSION_0_12_28',
  TerraformVersion_0_12_29 = 'TERRAFORM_VERSION_0_12_29',
  TerraformVersion_0_12_30 = 'TERRAFORM_VERSION_0_12_30',
  TerraformVersion_0_12_31 = 'TERRAFORM_VERSION_0_12_31',
  TerraformVersion_0_13_1 = 'TERRAFORM_VERSION_0_13_1',
  TerraformVersion_0_13_2 = 'TERRAFORM_VERSION_0_13_2',
  TerraformVersion_0_13_3 = 'TERRAFORM_VERSION_0_13_3',
  TerraformVersion_0_13_4 = 'TERRAFORM_VERSION_0_13_4',
  TerraformVersion_0_13_5 = 'TERRAFORM_VERSION_0_13_5',
  TerraformVersion_0_13_6 = 'TERRAFORM_VERSION_0_13_6',
  TerraformVersion_0_13_7 = 'TERRAFORM_VERSION_0_13_7',
  TerraformVersion_0_14_1 = 'TERRAFORM_VERSION_0_14_1',
  TerraformVersion_0_14_2 = 'TERRAFORM_VERSION_0_14_2',
  TerraformVersion_0_14_3 = 'TERRAFORM_VERSION_0_14_3',
  TerraformVersion_0_14_4 = 'TERRAFORM_VERSION_0_14_4',
  TerraformVersion_0_14_5 = 'TERRAFORM_VERSION_0_14_5',
  TerraformVersion_0_14_6 = 'TERRAFORM_VERSION_0_14_6',
  TerraformVersion_0_14_7 = 'TERRAFORM_VERSION_0_14_7',
  TerraformVersion_0_14_8 = 'TERRAFORM_VERSION_0_14_8',
  TerraformVersion_0_14_9 = 'TERRAFORM_VERSION_0_14_9',
  TerraformVersion_0_14_10 = 'TERRAFORM_VERSION_0_14_10',
  TerraformVersion_0_14_11 = 'TERRAFORM_VERSION_0_14_11',
  TerraformVersion_0_15_1 = 'TERRAFORM_VERSION_0_15_1',
  TerraformVersion_0_15_2 = 'TERRAFORM_VERSION_0_15_2',
  TerraformVersion_0_15_3 = 'TERRAFORM_VERSION_0_15_3',
  TerraformVersion_0_15_4 = 'TERRAFORM_VERSION_0_15_4',
  TerraformVersion_0_15_5 = 'TERRAFORM_VERSION_0_15_5',
  TerraformVersion_1_0_1 = 'TERRAFORM_VERSION_1_0_1',
  TerraformVersion_1_0_2 = 'TERRAFORM_VERSION_1_0_2',
  TerraformVersion_1_0_3 = 'TERRAFORM_VERSION_1_0_3',
  TerraformVersion_1_0_4 = 'TERRAFORM_VERSION_1_0_4',
  TerraformVersion_1_0_5 = 'TERRAFORM_VERSION_1_0_5',
  TerraformVersion_1_0_6 = 'TERRAFORM_VERSION_1_0_6',
  TerraformVersion_1_0_7 = 'TERRAFORM_VERSION_1_0_7',
  TerraformVersion_1_0_8 = 'TERRAFORM_VERSION_1_0_8',
  TerraformVersion_1_0_9 = 'TERRAFORM_VERSION_1_0_9',
  TerraformVersion_1_0_10 = 'TERRAFORM_VERSION_1_0_10',
  TerraformVersion_1_0_11 = 'TERRAFORM_VERSION_1_0_11',
  TerraformVersion_1_2_1 = 'TERRAFORM_VERSION_1_2_1',
  TerraformVersion_1_2_2 = 'TERRAFORM_VERSION_1_2_2',
  TerraformVersion_1_2_3 = 'TERRAFORM_VERSION_1_2_3',
  TerraformVersion_1_2_4 = 'TERRAFORM_VERSION_1_2_4',
  TerraformVersion_1_2_5 = 'TERRAFORM_VERSION_1_2_5',
  TerraformVersion_1_2_6 = 'TERRAFORM_VERSION_1_2_6',
  TerraformVersion_1_2_7 = 'TERRAFORM_VERSION_1_2_7',
  TerraformVersion_1_2_8 = 'TERRAFORM_VERSION_1_2_8',
  TerraformVersion_1_2_9 = 'TERRAFORM_VERSION_1_2_9',
  TerraformVersion_1_3_1 = 'TERRAFORM_VERSION_1_3_1',
  TerraformVersion_1_3_2 = 'TERRAFORM_VERSION_1_3_2',
  TerraformVersion_1_3_3 = 'TERRAFORM_VERSION_1_3_3',
  TerraformVersion_1_3_4 = 'TERRAFORM_VERSION_1_3_4',
  TerraformVersion_1_3_5 = 'TERRAFORM_VERSION_1_3_5',
  TerraformVersion_1_3_6 = 'TERRAFORM_VERSION_1_3_6',
  TerraformVersion_1_3_7 = 'TERRAFORM_VERSION_1_3_7',
  TerraformVersion_1_3_8 = 'TERRAFORM_VERSION_1_3_8',
  TerraformVersion_1_3_9 = 'TERRAFORM_VERSION_1_3_9',
  TerraformVersion_1_4_1 = 'TERRAFORM_VERSION_1_4_1',
  TerraformVersion_1_4_2 = 'TERRAFORM_VERSION_1_4_2',
  TerraformVersion_1_4_3 = 'TERRAFORM_VERSION_1_4_3',
  TerraformVersion_1_4_4 = 'TERRAFORM_VERSION_1_4_4',
  TerraformVersion_1_4_5 = 'TERRAFORM_VERSION_1_4_5',
  TerraformVersion_1_4_6 = 'TERRAFORM_VERSION_1_4_6',
  TerraformVersionLatest = 'TERRAFORM_VERSION_LATEST',
  TerraformVersionUnspecified = 'TERRAFORM_VERSION_UNSPECIFIED'
}

/** Represents the data about a Org member's Nuon account */
export type User = {
  __typename?: 'User';
  id: Scalars['ID']['output'];
};

/** Represents version control settings for the component */
export type VcsConfig = ConnectedGithubConfig | PublicGitConfig;

export type VcsConfigInput = {
  connectedGithub?: InputMaybe<ConnectedGithubConfigInput>;
  publicGit?: InputMaybe<PublicGitConfigInput>;
};
