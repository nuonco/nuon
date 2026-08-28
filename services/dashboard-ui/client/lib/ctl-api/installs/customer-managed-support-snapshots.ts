import { api } from '@/lib/api'
import type { TInstallationRegistration } from './register-customer-managed-install'

export type TCustomerManagedSupportSnapshotSummary = {
  id: string
  install_id: string
  created_at: string
  captured_at: string
  archive_sha256: string
  archive_size: number
  schema_version: number
  integrity_status: string
  association_status: string
  manifest: {
    schema_version: number
    captured_at: string
    registration_id: string
    bundle_digest: string
    producer: {
      name: string
      version?: string
      runner_version?: string
    }
  }
}

export type TCustomerManagedBundleContent = {
  kind: string
  name: string
  detail?: string
  digest?: string
  config_digest?: string
  size?: number
  component_definition?: Record<string, unknown>
  action_definition?: TCustomerManagedActionDefinition
  runbook_definition?: TCustomerManagedRunbookDefinition
}

export type TCustomerManagedActionStep = {
  name?: string
  command?: string
  inline_contents_digest?: string
  artifact_digest?: string
  index?: number
  environment?: Record<string, string>
}

export type TCustomerManagedActionTrigger = {
  type: string
  index?: number
  cron_schedule?: string
  component_name?: string
}

export type TCustomerManagedActionDefinition = {
  timeout_nanos?: number
  role?: string
  break_glass_role_arn?: string
  enable_kube_config?: boolean
  kubernetes_context_name?: string
  component_dependencies?: string[]
  references?: string[]
  triggers?: TCustomerManagedActionTrigger[]
  steps?: TCustomerManagedActionStep[]
}

export type TCustomerManagedRunbookStep = {
  kind?: string
  name?: string
  index?: number
  reference?: string
  component?: string
  role?: string
  plan_only?: boolean
  deploy_dependents?: boolean
  tear_down_dependents?: boolean
  skip_component_deploys?: boolean
  command?: string
  inline_contents_digest?: string
  environment?: Record<string, string>
  timeout_nanos?: number
  event_types?: string[]
  trigger_name?: string
}

export type TCustomerManagedRunbookDefinition = {
  readme_digest?: string
  inputs?: Array<{
    name: string
    display_name?: string
    description?: string
    default?: unknown
    type?: string
    index?: number
    required?: boolean
    sensitive?: boolean
  }>
  steps?: TCustomerManagedRunbookStep[]
}

export type TCustomerManagedBundleInfo = {
  schema_version: number
  deployment_id: string
  bundle_digest: string
  release?: { id: string; digest: string }
  package?: { id: string; digest: string; format: string; target: string }
  archive_digest?: string
  activated_at: string
  total_size?: number
  target?: { os: string; architecture: string }
  contents?: TCustomerManagedBundleContent[]
}

export type TCustomerManagedBundleChange = {
  kind: string
  name: string
  detail?: string
  change: 'added' | 'changed' | 'removed' | 'unchanged'
  previous_digest?: string
  candidate_digest?: string
  previous_config_digest?: string
  candidate_config_digest?: string
  previous_component_definition?: Record<string, unknown>
  candidate_component_definition?: Record<string, unknown>
  previous_action_definition?: TCustomerManagedActionDefinition
  candidate_action_definition?: TCustomerManagedActionDefinition
  previous_runbook_definition?: TCustomerManagedRunbookDefinition
  candidate_runbook_definition?: TCustomerManagedRunbookDefinition
}

export type TCustomerManagedBundleCandidate = {
  schema_version: number
  previous_digest: string
  staged_at: string
  archive_name?: string
  archive_size?: number
  bundle: TCustomerManagedBundleInfo
  changes: TCustomerManagedBundleChange[]
}

export type TCustomerManagedSnapshotRunStep = {
  id: string
  name: string
  kind: string
  job_id?: string
  status: string
  error?: string
  started_at?: string
  finished_at?: string
  source_run_id?: string
  result_directive?: string
  status_description?: string
  plan?: {
    kind: string
    content: unknown
  }
  drift?: Record<string, unknown>
}

export type TCustomerManagedSnapshotRun = {
  run_id: string
  dispatch_id?: string
  ref_id: string
  ref_kind: string
  ref_name: string
  source: string
  status: string
  error?: string
  bundle_digest?: string
  previous_run_id?: string
  started_at: string
  finished_at?: string
  steps: TCustomerManagedSnapshotRunStep[]
  result_directive?: string
}

export type TCustomerManagedSnapshotLogEntry = {
  time?: string
  level?: string
  msg?: string
  fields?: Record<string, unknown>
}

export type TCustomerManagedSnapshotJobLog = {
  job_id: string
  run_id?: string
  name?: string
  status?: string
  started_at?: string
  entries: TCustomerManagedSnapshotLogEntry[]
  total: number
  truncated?: boolean
}

export type TCustomerManagedSupportHealthTransition = {
  component_id?: string
  component_name?: string
  from: string
  to: string
  observed_at: string
}

export type TCustomerManagedCapturedInput = {
  name: string
  type?: string
  description?: string
  required?: boolean
  secret?: boolean
  bindable?: boolean
  value?: string
  default?: string
  value_status:
    | 'provided'
    | 'default'
    | 'redacted'
    | 'embedded-in-bundle'
    | 'unavailable'
  value_available?: boolean
}

export type TCustomerManagedCapturedRole = {
  name: string
  type: string
  cloud_id: string
  provisioned: boolean
}

export type TCustomerManagedSupportSnapshotData = {
  schema_version: number
  captured_at: string
  registration: TInstallationRegistration
  include_state?: boolean
  state?: {
    status?: Record<string, unknown>
    report?: Record<string, unknown>
  }
  runner?: {
    runner_id: string
    session_id: string
    version: string
    bundle_digest: string
    capabilities?: string[]
    started_at: string
    observed_at: string
  }
  catalog?: {
    deployment_id: string
    bundle_digest: string
    generated_at: string
    refs: Array<{
      id: string
      kind: string
      name: string
      component?: string
      steps?: number
    }>
  }
  active_bundle?: TCustomerManagedBundleInfo
  staged_bundle?: TCustomerManagedBundleCandidate
  bundle_history?: TCustomerManagedBundleInfo[]
  health?: {
    observed_at: string
    kind?: string
    cluster_access_error?: string
    components?: Array<{
      install_component_id?: string
      component_id?: string
      component_name?: string
      component_type?: string
      health: string
      truncated?: boolean
      resources?: Array<{
        kind?: string
        api_group?: string
        name?: string
        namespace?: string
        provider?: string
        health?: string
        message?: string
      }>
    }>
    sandbox_releases?: Array<{
      release_name: string
      namespace?: string
      health: string
      resources?: Array<{
        kind?: string
        api_group?: string
        name?: string
        namespace?: string
        provider?: string
        health?: string
        message?: string
      }>
    }>
  }
  health_transitions?: TCustomerManagedSupportHealthTransition[]
  current_inputs?: {
    observed_at: string
    inputs: TCustomerManagedCapturedInput[]
  }
  roles?: {
    observed_at: string
    roles: TCustomerManagedCapturedRole[]
  }
  runs?: TCustomerManagedSnapshotRun[]
  logs?: TCustomerManagedSnapshotJobLog[]
  collection: {
    schema_version: number
    redaction_policy: string
    included: string[]
    unavailable?: Record<string, string>
    truncated?: Record<string, number>
  }
}

export type TCustomerManagedSupportSnapshot =
  TCustomerManagedSupportSnapshotSummary & {
    snapshot: TCustomerManagedSupportSnapshotData
  }

export const getCustomerManagedSupportSnapshots = ({
  orgId,
  installId,
}: {
  orgId: string
  installId: string
}) =>
  api<TCustomerManagedSupportSnapshotSummary[]>({
    orgId,
    path: `installs/${installId}/support-snapshots`,
  })

export const getCustomerManagedSupportSnapshot = ({
  orgId,
  installId,
  snapshotId,
}: {
  orgId: string
  installId: string
  snapshotId: string
}) =>
  api<TCustomerManagedSupportSnapshot>({
    orgId,
    path: `installs/${installId}/support-snapshots/${snapshotId}`,
  })

export const uploadCustomerManagedSupportSnapshot = ({
  orgId,
  installId,
  file,
}: {
  orgId: string
  installId: string
  file: File
}) =>
  api<TCustomerManagedSupportSnapshot>({
    abortTimeout: 120000,
    body: file,
    headers: { 'Content-Type': 'application/octet-stream' },
    method: 'POST',
    orgId,
    path: `installs/${installId}/support-snapshots`,
  })
