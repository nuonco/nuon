/* eslint-disable */
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  /** Represents an ISO 8601-encoded date and time string. For example, 3:50 pm on September 7, 2019 in the time zone of UTC (Coordinated Universal Time) is represented as '2019-09-07T15:50:00Z' */
  DateTime: any;
};

/** Represents the AWS Authentication configuration used to access the vendor's OCI image */
export type AwsAuthConfig = {
  __typename?: 'AWSAuthConfig';
  /** AWS Region */
  region: AwsRegion;
  /** AWS IAM Role */
  role: Scalars['String'];
};

export type AwsAuthConfigInput = {
  region: AwsRegion;
  role: Scalars['String'];
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
  role: Scalars['String'];
};

export type AwsSettingsInput = {
  region: AwsRegion;
  role: Scalars['String'];
};

/** Represents a collection of general settings and information about a User's software */
export type App = Node & {
  __typename?: 'App';
  components: ComponentConnection;
  createdAt: Scalars['DateTime'];
  deployments: DeploymentConnection;
  id: Scalars['ID'];
  installs: InstallConnection;
  name: Scalars['String'];
  updatedAt: Scalars['DateTime'];
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
  totalCount: Scalars['Int'];
};

/** An auto-generated type which holds one App and a cursor during pagination */
export type AppEdge = {
  __typename?: 'AppEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of AppEdge */
  node: App;
};

export type AppInput = {
  githubInstallId?: InputMaybe<Scalars['ID']>;
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
  orgId?: InputMaybe<Scalars['ID']>;
};

/** Represents a basic Kubernetes deployment configurations */
export type BasicDeployConfig = {
  __typename?: 'BasicDeployConfig';
  /** The health check endpoint for the Kubernetes liveness probe */
  healthCheckPath: Scalars['String'];
  /** How many instances of deployment to maintain */
  instanceCount: Scalars['Int'];
  /** Container port */
  port: Scalars['Int'];
};

export type BasicDeployConfigInput = {
  healthCheckPath: Scalars['String'];
  instanceCount: Scalars['Int'];
  port: Scalars['Int'];
};

/** Represents information about a build */
export type Build = {
  __typename?: 'Build';
  componentId: Scalars['ID'];
  createdAt: Scalars['DateTime'];
  gitRef: Scalars['String'];
  id: Scalars['ID'];
  updatedAt: Scalars['DateTime'];
};

export type BuildConfigInput = {
  dockerBuildConfig?: InputMaybe<DockerBuildInput>;
  externalImageConfig?: InputMaybe<ExternalImageInput>;
  noop?: InputMaybe<Scalars['Boolean']>;
};

export type BuildInput = {
  componentId: Scalars['ID'];
  gitRef?: InputMaybe<Scalars['String']>;
};

/** Represents a collection of general settings and information about a piece of a App */
export type Component = Node & {
  __typename?: 'Component';
  app?: Maybe<App>;
  config?: Maybe<ComponentConfig>;
  createdAt: Scalars['DateTime'];
  deployments: DeploymentConnection;
  id: Scalars['ID'];
  name: Scalars['String'];
  updatedAt: Scalars['DateTime'];
};


/** Represents a collection of general settings and information about a piece of a App */
export type ComponentDeploymentsArgs = {
  options?: InputMaybe<ConnectionOptions>;
};

/** Represents all the component build configurations */
export type ComponentBuildConfig = DockerBuildConfig | ExternalImageConfig | NoopConfig;

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
  totalCount: Scalars['Int'];
};

/** Represents all the component deployment configurations */
export type ComponentDeployConfig = BasicDeployConfig | HelmRepoDeployConfig | NoopConfig;

/** An auto-generated type which holds one Component and a cursor during pagination */
export type ComponentEdge = {
  __typename?: 'ComponentEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of ComponentEdge */
  node?: Maybe<Component>;
};

export type ComponentInput = {
  appId?: InputMaybe<Scalars['ID']>;
  config?: InputMaybe<ComponentConfigInput>;
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
};

/** Represents settings for a connected github repository */
export type ConnectedGithubConfig = {
  __typename?: 'ConnectedGithubConfig';
  branch?: Maybe<Scalars['String']>;
  directory: Scalars['String'];
  gitRef?: Maybe<Scalars['String']>;
  repo: Scalars['String'];
};

export type ConnectedGithubConfigInput = {
  branch: Scalars['String'];
  directory: Scalars['String'];
  gitRef?: InputMaybe<Scalars['String']>;
  repo: Scalars['String'];
};

export type Connection = {
  /** Information to aid in pagination of list */
  pageInfo: PageInfo;
  /** Total count of items */
  totalCount: Scalars['Int'];
};

export type ConnectionOptions = {
  /** Returns the elements that come after the specified cursor */
  after?: InputMaybe<Scalars['String']>;
  /** Returns the elements that come before the specified cursor */
  before?: InputMaybe<Scalars['String']>;
  /** Returns first (n) elements */
  first?: InputMaybe<Scalars['Int']>;
  /** Returns last (n) elements */
  last?: InputMaybe<Scalars['Int']>;
};

/** Represents information about a deploy */
export type Deploy = {
  __typename?: 'Deploy';
  buildId: Scalars['ID'];
  createdAt: Scalars['DateTime'];
  id: Scalars['ID'];
  installId: Scalars['ID'];
  updatedAt: Scalars['DateTime'];
};

export type DeployConfigInput = {
  basicDeployConfig?: InputMaybe<BasicDeployConfigInput>;
  helmRepoDeployConfig?: InputMaybe<HelmRepoDeployConfigInput>;
  noop?: InputMaybe<Scalars['Boolean']>;
};

export type DeployInput = {
  buildId: Scalars['ID'];
  installIds?: InputMaybe<Array<Scalars['ID']>>;
};

/** Represents a collection of general settings and information about a deployed piece of an App */
export type Deployment = Node & {
  __typename?: 'Deployment';
  commitAuthor?: Maybe<Scalars['String']>;
  commitHash?: Maybe<Scalars['String']>;
  createdAt: Scalars['DateTime'];
  id: Scalars['ID'];
  updatedAt: Scalars['DateTime'];
};

/** An auto-generated type for paginating through multiple Deployments */
export type DeploymentConnection = Connection & {
  __typename?: 'DeploymentConnection';
  /** A list of edges */
  edges?: Maybe<Array<DeploymentEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int'];
};

/** An auto-generated type which holds one Deployment and a cursor during pagination */
export type DeploymentEdge = {
  __typename?: 'DeploymentEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of DeploymentEdge */
  node?: Maybe<Deployment>;
};

/** Represents a Docker build configuration */
export type DockerBuildConfig = {
  __typename?: 'DockerBuildConfig';
  /** Docker build arguments */
  buildArgs?: Maybe<Array<KeyValuePair>>;
  /** Name of Dockerfile used to build image */
  dockerfile: Scalars['String'];
  /** Version control system configuration */
  vcsConfig?: Maybe<VcsConfig>;
};

export type DockerBuildInput = {
  dockerfile: Scalars['String'];
  vcsConfig: VcsConfigInput;
};

/** Represents an external OCI image configuration */
export type ExternalImageConfig = {
  __typename?: 'ExternalImageConfig';
  authConfig?: Maybe<AwsAuthConfig>;
  /** URL where the OCI image is hosted */
  ociImageUrl: Scalars['String'];
  /** OCI image tag */
  tag: Scalars['String'];
};

export type ExternalImageInput = {
  authConfig?: InputMaybe<AwsAuthConfigInput>;
  ociImageUrl: Scalars['String'];
  tag?: InputMaybe<Scalars['String']>;
};

/** Represents a settings for GCP cloud target */
export type GcpSettings = {
  __typename?: 'GCPSettings';
  noop: Scalars['Boolean'];
};

export type GcpSettingsInput = {
  bogus: Scalars['String'];
};

/** Represents a public helm chart deployment configuration */
export type HelmRepoDeployConfig = {
  __typename?: 'HelmRepoDeployConfig';
  chartName: Scalars['String'];
  chartRepo: Scalars['String'];
  chartVersion: Scalars['String'];
  imageRepoValuesKey?: Maybe<Scalars['String']>;
  imageTagValuesKey?: Maybe<Scalars['String']>;
};

export type HelmRepoDeployConfigInput = {
  chartName: Scalars['String'];
  chartRepo: Scalars['String'];
  chartVersion: Scalars['String'];
  imageRepoValuesKey?: InputMaybe<Scalars['String']>;
  imageTagValuesKey?: InputMaybe<Scalars['String']>;
};

/** Represents a collection of general settings and information about a cloud target */
export type Install = Node & {
  __typename?: 'Install';
  components: ComponentConnection;
  createdAt: Scalars['DateTime'];
  deployments: DeploymentConnection;
  id: Scalars['ID'];
  name: Scalars['String'];
  settings: InstallSettings;
  updatedAt: Scalars['DateTime'];
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
  totalCount: Scalars['Int'];
};

/** An auto-generated type which holds one Install and a cursor during pagination */
export type InstallEdge = {
  __typename?: 'InstallEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of InstallEdge */
  node: Install;
};

export type InstallInput = {
  appId?: InputMaybe<Scalars['ID']>;
  awsSettings?: InputMaybe<AwsSettingsInput>;
  gcpSettings?: InputMaybe<GcpSettingsInput>;
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
};

/** Represents cloud target settings for AWS or GCP */
export type InstallSettings = AwsSettings | GcpSettings;

/** Represents a collection of general info about deployed piece of software on an Install */
export type Instance = {
  __typename?: 'Instance';
  hostname?: Maybe<Scalars['String']>;
  status: Status;
};

/** Represents a key value pair */
export type KeyValuePair = {
  __typename?: 'KeyValuePair';
  key: Scalars['String'];
  value: Scalars['String'];
};

export type Mutation = {
  __typename?: 'Mutation';
  cancelBuild: Scalars['Boolean'];
  createDeployment?: Maybe<Deployment>;
  deleteApp: Scalars['Boolean'];
  deleteComponent: Scalars['Boolean'];
  deleteInstall: Scalars['Boolean'];
  deleteOrg: Scalars['Boolean'];
  deleteSecrets: Scalars['Boolean'];
  echo: Scalars['String'];
  startBuild: Build;
  startDeploy: Deploy;
  upsertApp?: Maybe<App>;
  upsertComponent?: Maybe<Component>;
  upsertInstall?: Maybe<Install>;
  upsertOrg?: Maybe<Org>;
  upsertSecrets: Scalars['Boolean'];
};


export type MutationCancelBuildArgs = {
  id: Scalars['ID'];
};


export type MutationCreateDeploymentArgs = {
  componentId: Scalars['ID'];
};


export type MutationDeleteAppArgs = {
  id: Scalars['ID'];
};


export type MutationDeleteComponentArgs = {
  id: Scalars['ID'];
};


export type MutationDeleteInstallArgs = {
  id: Scalars['ID'];
};


export type MutationDeleteOrgArgs = {
  id: Scalars['ID'];
};


export type MutationDeleteSecretsArgs = {
  input?: InputMaybe<Array<SecretsIdsInput>>;
};


export type MutationEchoArgs = {
  word: Scalars['String'];
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
  createdAt: Scalars['DateTime'];
  /** A globally-unique identifier */
  id: Scalars['ID'];
  /** The date and time (ISO 8601 format) when the node was last updated */
  updatedAt: Scalars['DateTime'];
};

/** Represents no operation configuration */
export type NoopConfig = {
  __typename?: 'NoopConfig';
  noop: Scalars['Boolean'];
};

export type NoopSettings = {
  __typename?: 'NoopSettings';
  noop?: Maybe<Scalars['Boolean']>;
};

/** Represents a collection of general settings and information about a Nuon tenant */
export type Org = Node & {
  __typename?: 'Org';
  apps: AppConnection;
  createdAt: Scalars['DateTime'];
  githubInstallId?: Maybe<Scalars['ID']>;
  id: Scalars['ID'];
  name: Scalars['String'];
  updatedAt: Scalars['DateTime'];
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
  totalCount: Scalars['Int'];
};

/** An auto-generated type which holds one Org and a cursor during pagination */
export type OrgEdge = {
  __typename?: 'OrgEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of OrgEdge */
  node: Org;
};

export type OrgInput = {
  githubInstallId?: InputMaybe<Scalars['ID']>;
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
  ownerId?: InputMaybe<Scalars['ID']>;
};

/** Returns information about pagination in a connection, in accordance with the Relay specification */
export type PageInfo = {
  __typename?: 'PageInfo';
  /** The cursor corresponding to the last node in edges */
  endCursor?: Maybe<Scalars['String']>;
  /** Whether there are more pages to fetch following the current page */
  hasNextPage: Scalars['Boolean'];
  /** Whether there are any pages prior to the current page */
  hasPreviousPage: Scalars['Boolean'];
  /** The cursor corresponding to the first node in edges */
  startCursor?: Maybe<Scalars['String']>;
};

/** Represents settings for a public git repository */
export type PublicGitConfig = {
  __typename?: 'PublicGitConfig';
  directory: Scalars['String'];
  gitRef?: Maybe<Scalars['String']>;
  repo: Scalars['String'];
};

export type PublicGitConfigInput = {
  directory: Scalars['String'];
  gitRef?: InputMaybe<Scalars['String']>;
  repo: Scalars['String'];
};

export type Query = {
  __typename?: 'Query';
  app?: Maybe<App>;
  apps: AppConnection;
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
  instance: Instance;
  me?: Maybe<User>;
  org?: Maybe<Org>;
  orgStatus: Status;
  orgs: OrgConnection;
  ping?: Maybe<Scalars['String']>;
  repos: RepoConnection;
  secrets?: Maybe<Array<Secret>>;
};


export type QueryAppArgs = {
  id: Scalars['ID'];
};


export type QueryAppsArgs = {
  options?: InputMaybe<ConnectionOptions>;
  orgId: Scalars['ID'];
};


export type QueryBuildsArgs = {
  componentId: Scalars['ID'];
};


export type QueryComponentArgs = {
  id: Scalars['ID'];
};


export type QueryComponentsArgs = {
  appId: Scalars['ID'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryDeployArgs = {
  id: Scalars['ID'];
};


export type QueryDeploymentArgs = {
  id: Scalars['ID'];
};


export type QueryDeploymentStatusArgs = {
  appId: Scalars['ID'];
  componentId: Scalars['ID'];
  deploymentId: Scalars['ID'];
  orgId: Scalars['ID'];
};


export type QueryDeploymentsArgs = {
  appIds?: InputMaybe<Array<InputMaybe<Scalars['ID']>>>;
  componentIds?: InputMaybe<Array<InputMaybe<Scalars['ID']>>>;
  installIds?: InputMaybe<Array<InputMaybe<Scalars['ID']>>>;
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryInstallArgs = {
  id: Scalars['ID'];
};


export type QueryInstallStatusArgs = {
  appId: Scalars['ID'];
  installId: Scalars['ID'];
  orgId: Scalars['ID'];
};


export type QueryInstallsArgs = {
  appId: Scalars['ID'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryInstanceArgs = {
  appId: Scalars['ID'];
  componentId: Scalars['ID'];
  deploymentId: Scalars['ID'];
  installId: Scalars['ID'];
  orgId: Scalars['ID'];
};


export type QueryOrgArgs = {
  id: Scalars['ID'];
};


export type QueryOrgStatusArgs = {
  id: Scalars['ID'];
};


export type QueryOrgsArgs = {
  memberId: Scalars['ID'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QueryReposArgs = {
  githubInstallId: Scalars['ID'];
  options?: InputMaybe<ConnectionOptions>;
};


export type QuerySecretsArgs = {
  appId: Scalars['ID'];
  componentId: Scalars['ID'];
  installId: Scalars['ID'];
  orgId: Scalars['ID'];
};

export type Repo = {
  __typename?: 'Repo';
  defaultBranch?: Maybe<Scalars['String']>;
  fullName?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  owner?: Maybe<Scalars['String']>;
  url?: Maybe<Scalars['String']>;
};

/** An auto-generated type for paginating through multiple Installs */
export type RepoConnection = Connection & {
  __typename?: 'RepoConnection';
  /** A list of edges */
  edges?: Maybe<Array<RepoEdge>>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int'];
};

/** An auto-generated type which holds one Install and a cursor during pagination */
export type RepoEdge = {
  __typename?: 'RepoEdge';
  /** A cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of RepoEdge */
  node: Repo;
};

/** Represents a secret key-value configuration */
export type Secret = {
  __typename?: 'Secret';
  id: Scalars['ID'];
  key: Scalars['String'];
  value: Scalars['String'];
};

export type SecretInput = {
  id?: InputMaybe<Scalars['ID']>;
  key: Scalars['String'];
  value: Scalars['String'];
};

export type SecretsIdsInput = {
  appId: Scalars['ID'];
  componentId: Scalars['ID'];
  installId: Scalars['ID'];
  orgId: Scalars['ID'];
  secretId: Scalars['ID'];
};

export type SecretsInput = {
  appId: Scalars['ID'];
  componentId: Scalars['ID'];
  installId: Scalars['ID'];
  orgId: Scalars['ID'];
  secrets?: InputMaybe<Array<SecretInput>>;
};

export enum Status {
  Active = 'ACTIVE',
  Error = 'ERROR',
  Provisioning = 'PROVISIONING',
  Unknown = 'UNKNOWN',
  Unspecified = 'UNSPECIFIED'
}

/** Represents the data about a Org member's Nuon account */
export type User = {
  __typename?: 'User';
  id: Scalars['ID'];
};

/** Represents version control settings for the component */
export type VcsConfig = ConnectedGithubConfig | PublicGitConfig;

export type VcsConfigInput = {
  connectedGithub?: InputMaybe<ConnectedGithubConfigInput>;
  publicGit?: InputMaybe<PublicGitConfigInput>;
};
