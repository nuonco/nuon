// Core entity types matching Go app.* models

export type TOrg = {
  id: string
  name: string
  tags: string[] | null
  labels?: Record<string, string>
  created_at: string
  updated_at: string
  app_count?: number
  install_count?: number
  custom_cert?: boolean
  status?: string
  status_description?: string
}

export type TUserJourneyStep = {
  name: string
  title: string
  complete: boolean
  completed_at?: string
  completion_method?: string
  completion_source?: string
  metadata?: Record<string, unknown>
}

export type TUserJourney = {
  name: string
  title: string
  steps: TUserJourneyStep[]
}

export type TAccount = {
  id: string
  email: string
  subject: string
  account_type: string
  created_at: string
  updated_at: string
  roles?: TRole[]
  org_ids?: string[]
  user_journeys?: TUserJourney[]
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
  cloud_platform: string
  runner_type: string
  runner_status: string
  runner_status_description: string
  sandbox_status: string
  sandbox_status_description: string
  composite_component_status: string
  composite_component_status_description: string
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
  status?: string
  created_at: string
  updated_at: string
  config_count?: number
  org?: TOrg
}

export type TAppConfig = {
  id: string
  app_id: string
  version: number
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
  max_depth: number
  max_in_flight: number
  idle_timeout: number
  metadata: Record<string, string> | null
  workflow: any
  status_v2?: any
  created_at: string
  updated_at: string
  emitters?: TQueueEmitter[]
}

export type TQueueStatus = {
  Ready: boolean
  Stopped: boolean
  Paused: boolean
  QueueDepthCount: number
  InFlightCount: number
  InFlight: string[]
}

export type TQueueDetailResponse = {
  queue: TQueue
  status: TQueueStatus | null
  signals: TQueueSignal[]
  in_flight_signals: TQueueSignal[]
  temporal_ui_url: string
}

export type TQueueEmitter = {
  id: string
  queue_id: string
  name: string
  description: string
  mode: string
  cron_schedule: string
  scheduled_at: string
  fired: boolean
  signal_type: string
  signal_template: any
  status: any
  last_emitted_at: string
  next_emit_at: string
  emit_count: number
  workflow: { id: string; namespace: string }
  owner_id: string
  owner_type: string
  org_id: string
  created_at: string
  updated_at: string
  created_by_id: string
}

export type TQueueSignal = {
  id: string
  type: string
  queue_id: string
  owner_id: string
  owner_type: string
  status: any
  enqueued: boolean
  execution_count: number
  created_at: string
  updated_at: string
}

export type TWorkflow = {
  id: string
  type: string
  owner_id: string
  owner_type: string
  status: any
  created_at: string
  created_by_id: string
  started_at: string
  finished_at: string
  execution_time: number
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
  id: string
  timestamp: string
  severity_text: string
  severity_number: number
  body: string
  service_name: string
  trace_id: string
  span_id: string
  scope_name: string
  scope_version: string
  resource_schema_url: string
  scope_schema_url: string
  resource_attributes: Record<string, string>
  scope_attributes: Record<string, string>
  log_attributes: Record<string, string>
  runner_job_id: string
  runner_job_execution_step: string
  log_stream_id: string
  org_id: string
}

export type TSandboxModeJobConfig = {
  id: string
  job_type: string
  duration: number
  should_error: boolean
  error_message?: string
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
  entity_type: string
  entity_id: string
  entity_name: string
  created_at: string
  org_id: string | null
  org_name: string | null
  app_id: string | null
  app_name: string | null
  description: string | null
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

export type TLabelSearchResult = {
  entity_type: string
  entity_id: string
  entity_name: string
  labels: Record<string, string>
  detail_url: string
}

export type TOrgOption = {
  id: string
  name: string
}

export type TLabelsResponse = {
  results: TLabelSearchResult[]
  all_keys: string[]
  orgs: TOrgOption[]
  page: number
  total_pages: number
  total_count: number
}

export type TLogStreamLogsResponse = {
  logs: TLogEntry[]
  page: number
  total_pages: number
}

export type TOrgDetailResponse = {
  org: TOrg
  installs: TInstall[]
  recent_app: TApp | null
  graph_dot: string
  app_url: string
  page: number
  installs_total_pages: number
}

export type TAccountDetailResponse = {
  account: TAccount
  apps: TApp[]
  installs: TInstall[]
  audit_logs: TAuditLogEntry[]
  start_date: string
  end_date: string
  page: number
  installs_total_pages: number
  audit_logs_total_pages: number
}

export type TInstallDetailResponse = {
  install: TInstall
  active_deployments: any[]
  activity_logs: TAuditLogEntry[]
  start_date: string
  end_date: string
  app_url: string
  page: number
  activity_total_pages: number
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
  stacks?: any[]
  stack_config?: any
  templates: any[]
  flow_templates?: any[]
  all_signal_types?: string[]
  all_runner_job_types?: string[]
  all_runner_job_operation_types?: string[]
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

export type TAllRunnerView = {
  runner: TRunner
  org_name: string
  group_type: string
  process_online: boolean
  version: string
  process_type: string
  install_id: string
  install_name: string
}

export type TRunnerStatBucket = {
  label: string
  value: number
}

export type TAllRunnerStats = {
  group_type: TRunnerStatBucket[]
  version: TRunnerStatBucket[]
  process_type: TRunnerStatBucket[]
}

export type TAllRunnersResponse = {
  runners: TAllRunnerView[]
  orgs: TOrgOption[]
  stats: TAllRunnerStats
  page: number
  total_pages: number
  total_count: number
}

export type TInstallActivityResponse = {
  activity_logs: TAuditLogEntry[]
  audit_logs: TAuditLogEntry[]
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
