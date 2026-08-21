import type { TIconVariant } from '@/components/common/Icon'

export type TAPIError = {
  description: string
  error: string
  user_error: boolean
  meta?: any
  status?: number
}

export type TAPIResponse<T> = {
  data: T | null
  error: null | TAPIError
  headers: Record<string, string>
  status: Response['status']
}

export type TFileResponse = { content: string; filename: string }

export type TStepChangePlanType =
  | 'terraform_plan'
  | 'pulumi_plan'
  | 'helm_approval'
  | 'kubernetes_manifest_approval'
  | 'app_branch_plan'
  | 'install_creation'

export type TStepChangeStatus =
  | 'pending-approval'
  | 'approved'
  | 'denied'
  | 'applied'
  | 'generating'
  | 'error'

export type TStepChangeCounts = {
  create: number
  update: number
  delete: number
  replace: number
  noop: number
}

export type TStepChangeSummary = {
  stepId: string
  stepName: string
  componentName?: string
  planType: TStepChangePlanType
  status: TStepChangeStatus
  counts: TStepChangeCounts
  hasDetail: boolean
}

export type TPaginationPageData = {
  hasNext: string
  offset: string
}

export type TPaginationParams = {
  offset?: number | string
  limit?: number | string
}

export type TTheme =
  | 'default'
  | 'neutral'
  | 'info'
  | 'success'
  | 'warn'
  | 'brand'
  | 'error'

export type TNavLink = {
  badge?: boolean
  iconVariant?: TIconVariant
  path: string
  text: string
  isExternal?: boolean
  shortcut?: string
  matchPaths?: string[]
}

export type TNavSectionHeader = {
  type: 'section'
  label: string
}

export type TNavAction = {
  type: 'action'
  key: string
  iconVariant?: TIconVariant
  text: string
  onClick: () => void
  isActive?: boolean
}

export type TNavItem = TNavLink | TNavSectionHeader | TNavAction

export type TEmptyVariant =
  | '404'
  | 'actions'
  | 'app'
  | 'diagram'
  | 'history'
  | 'policy'
  | 'search'
  | 'table'

export type TKeyValue = {
  key: string
  value: string
  type?: string
}

export type TTerraformChangeAction =
  | 'create'
  | 'update'
  | 'delete'
  | 'no-op'
  | 'replace'
  | 'read'

export type TTerraformResourceChange = {
  address: string
  module?: string | null
  resource: string
  name: string
  action: TTerraformChangeAction
  before?: any
  after?: any
}

export type TTerraformOutputChange = {
  output: string
  action: TTerraformChangeAction
  before?: any
  after?: any
  afterUnknown?: any
  afterSensitive?: any
  beforeSensitive?: any
}

export type TTerraformPlan = {
  resource_drift?: Array<{
    address: string
    module_address?: string | null
    type: string
    name: string
    change: {
      actions: TTerraformChangeAction[]
      before?: any
      after?: any
      after_unknown?: any
    }
  }>
  resource_changes: Array<{
    address: string
    module_address?: string | null
    type: string
    name: string
    change: {
      actions: TTerraformChangeAction[]
      before?: any
      after?: any
      after_unknown?: any
    }
  }>
  output_changes?: {
    [name: string]: {
      actions: TTerraformChangeAction[]
      before?: any
      after?: any
      after_unknown?: any
      after_sensitive?: any
      before_sensitive?: any
    }
  }
}

export type TPulumiChangeAction =
  | 'create'
  | 'update'
  | 'delete'
  | 'replace'
  | 'create-replacement'
  | 'delete-replaced'
  | 'same'
  | 'read'
  | 'refresh'

export type THelmK8sChangeAction =
  | 'add'
  | 'added'
  | 'change'
  | 'changed'
  | 'destroy'
  | 'destroyed'

type TPlanSummary = {
  add: number
  change: number
  destroy: number
}

type TPlanChange = {
  resource: string
  resourceType: string
  action: THelmK8sChangeAction
  before?: string
  after?: string
  diff: string
}

export type THelmPlanSummary = TPlanSummary
export type TKubernetesPlanSummary = TPlanSummary
export type TKubernetesPlanChange = TPlanChange & {
  name: string
  namespace: string
}
export type THelmPlanChange = TPlanChange & {
  workspace: string
  release: string
}

export type TKubernetesPlanItem = {
  _version: string
  name: string
  namespace: string
  kind: string
  api: string
  resource: string
  op: string
  type: number // 1=add, 2=delete, 3=change
  dry_run: boolean
  error?: string
  entries?: Array<{
    path: string
    original: string
    applied: string
    type: number
    payload: string
  }>
}

export type TKubernetesPlan = {
  plan: string
  op: string
  k8s_content_diff: TKubernetesPlanItem[]
}

export type TKubernetesPlanError = {
  namespace: string
  name: string
  resource: string
  resourceType: string
  error: string
}

export type THelmPlan = {
  plan: string
  op: string
  helm_content_diff: {
    api: string
    kind: string
    name: string
    namespace: string
    before: string
    after: string
    entries?: Array<{
      path: string
      original: string
      applied: string
      type: number
      payload: string
    }>
  }[]
}

export type TCloudPlatform = 'aws' | 'azure' | 'gcp' | 'unknown'

export type TNuonVersion = {
  api: {
    git_ref: string
    version: string
  }
  ui: {
    version: string
  }
}

export type TAPIHealth = { status: 'ok' | 'degraded'; degraded: string[] }

export interface IUser {
  email?: string
  name?: string
  picture?: string
  sub?: string
  [key: string]: any
}
