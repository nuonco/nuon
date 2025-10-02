import type { ReactNode } from 'react'
import type { TIconVariant } from '@/components/common/Icon'

// TODO(nnnat): old types replace with types below
export type TRouteParams<S extends string | number | symbol = string> = Record<
  S,
  string
>
export type TRouteSearchParams<S extends string | number | symbol = string> =
  Record<S, string>

export interface IPageProps<
  P extends string | number | symbol = string,
  S extends string | number | symbol = string,
> {
  params?: Promise<TRouteParams<P>>
  searchParams?: Promise<TRouteSearchParams<S>>
}

export interface ILayoutProps<
  P extends string | number | symbol = string,
  S extends string | number | symbol = string,
> {
  children: ReactNode
  params?: Promise<TRouteParams<P>>
  searchParams?: Promise<TRouteSearchParams<S>>
}

export interface IRouteProps extends IPageProps {}
// -- end old types ---

// nextjs types
export type TParams<Keys extends string> = Promise<Record<Keys, string>>

export type TRouteProps<Keys extends string, T = {}> = {
  params: TParams<Keys>
} & T

export type TPageProps<Keys extends string, T = {}> = {
  params: TParams<Keys>
  searchParams: Promise<Record<string, string>>
} & T

export type TLayoutProps<Keys extends string, T = {}> = {
  children: ReactNode
  params: TParams<Keys>
} & T

// fetch wrapper types
export type TAPIError = {
  description: string
  error: string
  user_error: boolean
  meta?: any
}

export type TAPIResponse<T> = {
  data: T | null
  error: null | TAPIError
  headers: Record<string, string>
  status: Response['status']
}

export type TFileResponse = { content: string; filename: string }

export type TPaginationPageData = {
  hasNext: string
  offset: string
}

export type TPaginationParams = {
  offset?: number | string
  limit?: number | string
}

// page nav link types
export type TNavLink = {
  iconVariant?: TIconVariant
  path: string
  text: string
  isExternal?: boolean
}

// UI variant types
export type TEmptyVariant =
  | '404'
  | 'actions'
  | 'diagram'
  | 'history'
  | 'search'
  | 'table'

// Key value type
export type TKeyValue = {
  key: string
  value: string
  type?: string
}

// Terraform plan types
export type TTerraformChangeAction =
  | 'create'
  | 'update'
  | 'delete'
  | 'noop'
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

// Helm & K8s plan types
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
  group_version_kind: {
    Group: string
    Version: string
    Kind: string
  }
  group_version_resource: {
    Group: string
    Version: string
    Resource: string
  }
  namespace: string
  name: string
  before?: string
  after?: string
  op: string
}

export type TKubernetesPlan = TKubernetesPlanItem[]

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
  }[]
}

// cloud platform
export type TCloudPlatform = 'aws' | 'azure' | 'gcp' | 'unknown'
