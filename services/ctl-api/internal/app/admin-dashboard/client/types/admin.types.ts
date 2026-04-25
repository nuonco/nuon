// Core entity types matching Go app.* models

export type TOrg = {
  id: string
  name: string
  tags: string[] | null
  created_at: string
  updated_at: string
  app_count?: number
  install_count?: number
  custom_cert?: boolean
  status?: string
  status_description?: string
}

export type TAccount = {
  id: string
  email: string
  account_type: string
  created_at: string
  updated_at: string
  roles?: TRole[]
  org_ids?: string[]
}

export type TRole = {
  id: string
  role_type: string
  org_id: string
  org?: TOrg
  created_at: string
}

export type TInstall = {
  id: string
  name: string
  org_id: string
  app_id: string
  status: string
  status_description: string
  created_at: string
  updated_at: string
  deleted_at?: number
  created_by_id: string
  labels?: Record<string, string>
  org?: TOrg
  app?: TApp
  app_config?: TAppConfig
  app_runner_config?: TAppRunnerConfig
  runner_group?: TRunnerGroup
}

export type TApp = {
  id: string
  name: string
  org_id: string
  created_at: string
  updated_at: string
  config_count?: number
}

export type TAppConfig = {
  id: string
  app_id: string
  created_at: string
}

export type TAppRunnerConfig = {
  id: string
  app_id: string
  created_at: string
}

export type TRunnerGroup = {
  id: string
  install_id: string
  runners?: TRunner[]
}

export type TRunner = {
  id: string
  name: string
  display_name: string
  runner_group_id: string
  created_at: string
  updated_at: string
  status?: string
}

export type TRunnerProcess = {
  id: string
  runner_id: string
  status: string
  version: string
  created_at: string
}

export type TQueue = {
  id: string
  name: string
  owner_id: string
  owner_type: string
  workflow: any
  created_at: string
  emitters?: TQueueEmitter[]
}

export type TQueueEmitter = {
  id: string
  queue_id: string
  name: string
  owner_id: string
  owner_type: string
  created_at: string
}

export type TQueueSignal = {
  id: string
  type: string
  queue_id: string
  owner_id: string
  owner_type: string
  status: any
  created_at: string
  updated_at: string
}

export type TWorkflow = {
  id: string
  type: string
  owner_id: string
  owner_type: string
  status: string
  created_at: string
  created_by_id: string
  steps?: TWorkflowStep[]
}

export type TWorkflowStep = {
  id: string
  workflow_id: string
  step_target_id: string
  step_target_type: string
  group_idx: number
  group_retry_idx: number
  idx: number
  status: string
  created_at: string
  queue_signal?: TQueueSignal
  approval?: any
}

export type TWorkflowStepGroup = {
  group_idx: number
  group_retry_idx: number
  status: string
}

export type TLogStream = {
  id: string
  org_id: string
  owner_id: string
  owner_type: string
  created_at: string
}

export type TLogEntry = {
  timestamp: string
  severity_text: string
  body: string
  scope_name: string
  resource_attributes: Record<string, string>
  log_attributes: Record<string, string>
}

export type TSandboxModeJobConfig = {
  id: string
  job_type: string
  duration: number
  should_error: boolean
  panic: boolean
  trigger_shutdown: boolean
  created_at: string
  updated_at: string
}

export type TSandboxModeSignalConfig = {
  id: string
  signal_type: string
  frequency: string
  is_disabled: boolean
  created_at: string
  updated_at: string
}

export type TAuditLogEntry = {
  id: string
  entity_type: string
  entity_id: string
  action: string
  created_at: string
  account_id: string
  metadata: any
}

// Temporal-specific types

export type TWorkflowInfo = {
  status: string
  activities: TActivityInfo[]
  child_workflows: TChildWorkflowInfo[]
  awaited_signals: TAwaitedSignalInfo[]
  update_handlers: string[]
  update_executions: TUpdateExecution[]
  orphan_activities: TActivityInfo[]
}

export type TActivityInfo = {
  name: string
  status: string
  started_at: string
  finished_at: string
  duration: number
  attempt: number
  failure: string
  input: string
  result: string
  scheduled_event_id: number
}

export type TChildWorkflowInfo = {
  workflow_type: string
  workflow_id: string
  run_id: string
  namespace: string
  status: string
  started_at: string
  finished_at: string
  duration: number
  failure: string
}

export type TAwaitedSignalInfo = {
  queue_signal_id: string
  signal: TQueueSignal | null
  status: string
  started_at: string
  finished_at: string
  duration: number
  failure: string
}

export type TUpdateExecution = {
  name: string
  update_id: string
  status: string
  started_at: string
  finished_at: string
  duration: number
  input: string
  result: string
  failure: string
  activities: TActivityInfo[]
}

export type TNamespaceWorkerInfo = {
  namespace: string
  task_queue: string
  error: string
  workflow_pollers: TPollerDetail[]
  activity_pollers: TPollerDetail[]
  workflow_stats: TTaskQueueStatsInfo | null
  activity_stats: TTaskQueueStatsInfo | null
  total_poller_count: number
  is_healthy: boolean
}

export type TPollerDetail = {
  identity: string
  last_access_time: string
  rate_per_second: number
}

export type TTaskQueueStatsInfo = {
  approximate_backlog_count: number
  approximate_backlog_age: number
  tasks_add_rate: number
  tasks_dispatch_rate: number
}

export type TStepDetailData = {
  step: TWorkflowStep
  queue_signal_json: string
  step_target: TStepTargetData | null
}

export type TGroupDetailData = {
  group: TWorkflowStepGroup
  steps: TStepDetailData[]
}

export type TStepTargetData = {
  id: string
  type: string
  status: string
  log_stream_id: string
}

export type TRunnerDetailView = {
  runner: TRunner
  install_id: string
  install_name: string
  process: TRunnerProcess | null
  process_online: boolean
  configs: Record<string, TSandboxModeJobConfig>
}

export type TSandboxRunnerView = {
  runner: TRunner
  process_online: boolean
  version: string
  configs: TSandboxModeJobConfig[]
  install_id: string
  install_name: string
}

// API response types

export type TOrgsResponse = {
  orgs: TOrg[]
  all_tags: string[]
  page: number
  total_pages: number
}

export type TAccountsResponse = {
  accounts: TAccount[]
  page: number
  total_pages: number
}

export type TInstallsResponse = {
  installs: TInstall[]
  page: number
  total_pages: number
}

export type TQueuesResponse = {
  queues: TQueue[]
  page: number
  total_pages: number
}

export type TWorkflowsResponse = {
  workflows: TWorkflow[]
  page: number
  total_pages: number
}

export type TQueueSignalsResponse = {
  signals: TQueueSignal[]
  page: number
  total_pages: number
  signal_types?: string[]
}

export type TLabelsResponse = {
  labels: Record<string, string[]>
}

export type TLogStreamLogsResponse = {
  logs: TLogEntry[]
  page: number
  total_pages: number
}

export type TOrgDetailResponse = {
  org: TOrg
  installs: TInstall[]
  component_graph?: string
  support_users: string[]
  page: number
  total_pages: number
}

export type TAccountDetailResponse = {
  account: TAccount
  apps: TApp[]
  installs: TInstall[]
}

export type TInstallDetailResponse = {
  install: TInstall
}

export type TWorkflowDetailResponse = {
  workflow: TWorkflow
  groups: TGroupDetailData[]
  workflow_info: TWorkflowInfo | null
  temporal_workflow_id: string
  temporal_run_id: string
  temporal_namespace: string
}

export type TSandboxModeResponse = {
  runner_job_configs: TSandboxModeJobConfig[]
  signal_configs: TSandboxModeSignalConfig[]
  stacks: any[]
  templates: any[]
}

export type TTemporalWorkersResponse = {
  workers: TNamespaceWorkerInfo[]
}

export type TInFlightSignalsResponse = {
  signals: TQueueSignal[]
  page: number
  total_pages: number
}

export type TSignalCatalogResponse = {
  signal_types: string[]
}

export type TSignalCatalogDetailResponse = {
  signal_type: string
  recent_signals: TQueueSignal[]
}

export type TRunnersResponse = {
  runners: TSandboxRunnerView[]
}

export type TInstallActivityResponse = {
  entries: TAuditLogEntry[]
  page: number
  total_pages: number
}

export type TInstallActiveDeploymentsResponse = {
  deployments: any[]
}

export type TInstallWorkflowsResponse = {
  workflows: TWorkflow[]
  page: number
  total_pages: number
}
