import type {
  TAppReleaseWithFiles,
  TOTELLog,
  TWorkflow,
  TWorkflowStep,
} from "@/types/ctl-api.types";

export type TCatalogRef = {
  id: string;
  kind: string;
  name: string;
  component?: string;
  cron_schedule?: string;
  steps?: number;
};

export type TPortalMode = {
  mode: "offline" | "connected";
  capabilities: Record<string, boolean>;
};

export type TPortalBranding = {
  name: string;
  logo_url?: string;
  favicon_url?: string;
  primary_color: string;
  support_url?: string;
};

export type TConnectedRelease = {
  release: TAppReleaseWithFiles;
  active: boolean;
};

export type TConnectedReleaseUpdate = {
  id: string;
  app_release_id?: string;
  workflow_id?: string;
  created_at: string;
  status?: {
    status?: string;
    status_human_description?: string;
    metadata?: Record<string, unknown>;
  };
  workflow?: TConnectedWorkflow;
};

export type TConnectedApproval = Omit<
  NonNullable<TWorkflowStep["approval"]>,
  "id" | "type" | "response"
> & {
  id: string;
  type: NonNullable<NonNullable<TWorkflowStep["approval"]>["type"]>;
  response?: { type: string; note?: string };
};
export type TConnectedWorkflowStep = Omit<
  TWorkflowStep,
  "id" | "name" | "status" | "approval" | "log_stream"
> & {
  id: string;
  name: string;
  status?: TWorkflowStep["status"];
  approval?: TConnectedApproval;
  log_stream?: NonNullable<TWorkflowStep["log_stream"]> & { id: string };
};
export type TConnectedLog = TOTELLog & {
  id: string;
  timestamp: string;
  body: string;
};
export type TConnectedWorkflow = Omit<
  TWorkflow,
  "id" | "type" | "created_at" | "status" | "steps"
> & {
  id: string;
  name?: string;
  type: NonNullable<TWorkflow["type"]>;
  created_at: string;
  started_at?: string;
  finished_at?: string;
  status?: TWorkflow["status"];
  steps?: TConnectedWorkflowStep[];
};

export type TCatalog = {
  deployment_id: string;
  bundle_digest: string;
  generated_at: string;
  refs?: TCatalogRef[];
};

export type TResourceHealth = {
  kind?: string;
  api_group?: string;
  name?: string;
  namespace?: string;
  provider?: string;
  health?: string;
  message?: string;
};

export type TComponentHealth = {
  component_id?: string;
  install_component_id?: string;
  component_name?: string;
  component_type?: string;
  health: string;
  truncated?: boolean;
  resources?: TResourceHealth[];
};

export type THealthTransition = {
  component_id?: string;
  component_name?: string;
  from: string;
  to: string;
  observed_at: string;
};

export type THealth = {
  latest?: {
    observed_at?: string;
    components?: TComponentHealth[];
  } | null;
  transitions?: THealthTransition[];
};

export type TRunnerHeartbeat = {
  runner_id: string;
  session_id: string;
  version: string;
  bundle_digest: string;
  capabilities?: string[];
  started_at: string;
  observed_at: string;
};

export type TStatusStep = {
  id: string;
  name: string;
  status: string;
  started_at?: string;
  finished_at?: string;
  error?: string;
};

export type TStatus = {
  install_id: string;
  run_id: string;
  status: string;
  bundle_digest?: string;
  approval_required?: boolean;
  approval_phase?: string;
  failed_step?: string;
  started_at: string;
  finished_at?: string;
  heartbeat_at?: string;
  steps?: TStatusStep[];
};

export type TStepReport = {
  id: string;
  name: string;
  job_type: string;
  job_operation: string;
  job_group: string;
  status: string;
  started_at?: string;
  finished_at?: string;
  error?: string;
  executions: number;
  success?: boolean;
  error_code?: string;
  outputs?: unknown;
};

export type TReport = {
  install_id: string;
  run_id: string;
  status: string;
  failed_step?: string;
  started_at: string;
  finished_at?: string;
  steps?: TStepReport[];
};

export type TStackOutputs = Record<string, unknown>;

export type TInstallStackEvent = {
  id: string;
  logical_resource_id?: string;
  resource_type?: string;
  status: string;
  status_reason?: string;
  timestamp: string;
};

export type TInstallStack = {
  name: string;
  status: string;
  phase: "pending" | "in-progress" | "finished" | "failed";
  status_reason?: string;
  started_at?: string;
  updated_at?: string;
  events: TInstallStackEvent[];
  resources?: Record<
    string,
    {
      type: string;
      properties: Record<string, unknown>;
    }
  >;
};

export type TBundleContent = {
  kind: string;
  name: string;
  detail?: string;
  digest?: string;
  config_digest?: string;
  size?: number;
  component_definition?: Record<string, unknown>;
  action_definition?: TBundleActionDefinition;
  runbook_definition?: TBundleRunbookDefinition;
};

export type TBundleActionDefinition = {
  timeout_nanos?: number;
  role?: string;
  break_glass_role_arn?: string;
  enable_kube_config?: boolean;
  kubernetes_context_name?: string;
  component_dependencies?: string[];
  references?: string[];
  triggers?: TBundleActionTrigger[];
  steps?: TBundleActionStep[];
};

export type TBundleActionTrigger = {
  type: string;
  index?: number;
  cron_schedule?: string;
  component_name?: string;
};

export type TBundleActionStep = {
  name: string;
  command?: string;
  inline_contents_digest?: string;
  source?: TBundleSource;
  artifact_digest?: string;
  index?: number;
  environment?: Record<string, string>;
};

export type TBundleRunbookDefinition = {
  readme_digest?: string;
  inputs?: TBundleRunbookInput[];
  steps: TBundleRunbookStep[];
};

export type TBundleRunbookStep = {
  kind: string;
  name?: string;
  index?: number;
  reference?: string;
  component?: string;
  role?: string;
  plan_only?: boolean;
  deploy_dependents?: boolean;
  tear_down_dependents?: boolean;
  skip_component_deploys?: boolean;
  command?: string;
  inline_contents_digest?: string;
  environment?: Record<string, string>;
  timeout_nanos?: number;
  trigger_name?: string;
  event_types?: string[];
  filters_digest?: string;
};

export type TBundleRunbookInput = {
  name: string;
  display_name?: string;
  description?: string;
  default?: string;
  type?: string;
  index?: number;
  required?: boolean;
  sensitive?: boolean;
};

export type TBundleSource = {
  repository?: string;
  requested_ref?: string;
  commit?: string;
  directory?: string;
  version?: string;
  digest?: string;
};

export type TBundleInfo = {
  deployment_id: string;
  bundle_digest: string;
  release?: { id: string; digest: string };
  package?: { id: string; digest: string; format: string; target: string };
  archive_digest?: string;
  activated_at: string;
  target?: { os: string; architecture: string };
  verification: { blobs_verified: boolean; envelope_parsed: boolean };
  total_size?: number;
  contents?: TBundleContent[];
};

export type TBundle = {
  active: TBundleInfo | null;
  candidate: TBundleCandidate | null;
  candidate_record_key?: string;
  stack_candidate: TStackCandidate | null;
  history: TBundleInfo[];
  comparisons: TBundleHistoryComparison[];
};

export type TBundleHistoryComparison = {
  previous_digest: string;
  bundle_digest: string;
  available: boolean;
  changes?: TBundleChange[];
};

export type TStackCandidate = {
  schema_version: number;
  bundle_digest: string;
  candidate_record_key?: string;
  stack_name: string;
  change_set_name: string;
  status: string;
  execution_status: string;
  status_reason?: string;
  no_op?: boolean;
  changes?: TStackChange[];
  created_at: string;
  stack_applied_at?: string;
  runner_activated_at?: string;
  instance_refresh_id?: string;
};

export type TStackChange = {
  action: string;
  logical_resource_id: string;
  resource_type: string;
  replacement?: string;
  scope?: string[];
  details?: TStackChangeDetail[];
  property_changes?: TStackPropertyChange[];
  property_changes_captured?: boolean;
  property_changes_truncated?: boolean;
};

export type TStackPropertyChange = {
  path: string;
  before?: unknown;
  after?: unknown;
};

export type TStackChangeDetail = {
  attribute?: string;
  name?: string;
  requires_recreation?: string;
  evaluation?: string;
  change_source?: string;
  causing_entity?: string;
};

export type TBundleChange = {
  kind: string;
  name: string;
  detail?: string;
  change: "added" | "changed" | "removed" | "unchanged";
  previous_digest?: string;
  candidate_digest?: string;
  previous_config_digest?: string;
  candidate_config_digest?: string;
  previous_component_definition?: Record<string, unknown>;
  candidate_component_definition?: Record<string, unknown>;
  previous_action_definition?: TBundleActionDefinition;
  candidate_action_definition?: TBundleActionDefinition;
  previous_runbook_definition?: TBundleRunbookDefinition;
  candidate_runbook_definition?: TBundleRunbookDefinition;
  plan_step_id?: string;
  apply_step_id?: string;
};

export type TBundleCandidate = {
  schema_version: number;
  previous_digest: string;
  staged_at: string;
  archive_name?: string;
  archive_size?: number;
  bundle: TBundleInfo;
  changes: TBundleChange[];
  deployment?: {
    stack_template_url: string;
    candidate_bundle_key: string;
    target_bundle_key: string;
  };
};

export type TBundleUploadStatus = {
  state: "" | "uploading" | "processing" | "complete" | "failed";
  phase: string;
  detail: string;
  updated_at: string;
};

export type TPlanChange = {
  actions?: string[];
  before?: unknown;
  after?: unknown;
  after_unknown?: unknown;
};

export type TPlanResourceChange = {
  address: string;
  change?: TPlanChange;
};

export type TPlan = {
  resource_changes?: TPlanResourceChange[];
  resource_drift?: TPlanResourceChange[];
  output_changes?: Record<
    string,
    { actions?: string[]; before?: unknown; after?: unknown }
  >;
};

// The late-bound composite plan the customer-managed runner renders for a bootstrap
// step: top-level keys such as deploy_plan or sandbox_run_plan describe what
// the step's job will execute.
export type TCompositePlan = Record<string, unknown>;

export type TDiffEntry = {
  path?: string;
  original?: unknown;
  applied?: unknown;
  type: string | number;
  changes?: Record<string, unknown>;
  payload?: string;
};

export type TResourceDiff = {
  _version?: string;
  name?: string;
  namespace?: string;
  kind?: string;
  api?: string;
  resource?: string;
  op?: string;
  type?: string | number;
  error?: string;
  entries?: TDiffEntry[];
};

export type TManifestDiffContent = {
  plan?: string;
  op?: string;
  helm_content_diff?: TResourceDiff[];
  k8s_content_diff?: TResourceDiff[];
  template_output?: string;
  dry_run_output?: string;
};

// The decoded execution result a deploy handler persisted for a bootstrap
// step: the real `terraform plan` JSON, or helm/k8s manifest resource diffs.
export type TStepResult = {
  success: boolean;
  kind: "terraform" | "helm" | "kubernetes_manifest" | "unknown";
  content: (TPlan & TManifestDiffContent) | string | null;
};

export type TDriftResource = {
  address: string;
  action: string;
  drifted?: boolean;
};

export type TDrift = {
  drifted: boolean;
  resource_changes: number;
  output_changes: number;
  resource_drift: number;
  summary?: string;
  resources?: TDriftResource[];
  resources_truncated?: boolean;
};

export type TRunStep = {
  id: string;
  name: string;
  kind: string;
  job_id?: string;
  status: string;
  error?: string;
  started_at?: string;
  finished_at?: string;
  source_run_id?: string;
  result_directive?: string;
  status_description?: string;
  drift?: TDrift;
};

export type TJobLogSummary = {
  job_id: string;
  logs_available: boolean;
  source?: string;
  name?: string;
  status?: string;
  run_id?: string;
  ref_name?: string;
  started_at?: string;
};

export type TLogEntry = {
  time?: string;
  level?: string;
  msg?: string;
  fields?: Record<string, unknown>;
  raw?: string;
};

export type TJobLog = {
  job_id: string;
  total: number;
  truncated: boolean;
  entries: TLogEntry[];
};

export type TRun = {
  run_id: string;
  dispatch_id?: string;
  ref_id: string;
  ref_kind: string;
  ref_name: string;
  source: string;
  status: string;
  error?: string;
  bundle_digest?: string;
  previous_run_id?: string;
  started_at: string;
  finished_at?: string;
  result_directive?: string;
  events?: Array<{
    schema_version: number;
    sequence: number;
    created_at: string;
    status: { status: string; failed_step?: string; result_directive?: string };
  }>;
  steps?: TRunStep[];
};
