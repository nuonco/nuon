import { api } from '@/lib/api'

// All entity types intentionally use loose shapes — the BFF returns raw GORM
// models and only a handful of fields are needed to render the diagram. Add
// fields here as the diagram surfaces them.

export type TCompositeStatus = {
  status?: string
  status_human_description?: string
  created_at_ts?: number
}

export type TDataModelEntity = {
  id: string
  name?: string
  status?: TCompositeStatus
  status_v2?: TCompositeStatus
}

export type TDataModelOrg = TDataModelEntity

export type TDataModelApp = TDataModelEntity & {
  org_id: string
}

export type TDataModelComponent = TDataModelEntity & {
  org_id: string
  app_id: string
  type?: string
}

export type TDataModelInstall = TDataModelEntity & {
  org_id: string
  app_id: string
  runner_id?: string
}

export type TDataModelRunner = TDataModelEntity & {
  org_id: string
}

export type TDataModelWorkflow = {
  id: string
  org_id: string
  owner_id: string
  owner_type: string
  status?: TCompositeStatus
  created_at?: string
}

export type TDataModelStep = {
  id: string
  owner_id: string
  owner_type: string
  name: string
  idx: number
  status?: TCompositeStatus
}

export type TDataModelQueue = TDataModelEntity & {
  org_id?: string
  owner_id: string
  owner_type: string
  max_in_flight?: number
}

export type TDataModelEmitter = {
  id: string
  queue_id: string
  mode?: string
  status?: TCompositeStatus
}

export type TDataModelSignal = {
  id: string
  queue_id: string
  emitter_id?: string | null
  owner_id?: string
  owner_type?: string
  type?: string
  status?: TCompositeStatus
  created_at?: string
}

export type TDataModelResponse = {
  org: TDataModelOrg | null
  apps: TDataModelApp[]
  components: TDataModelComponent[]
  installs: TDataModelInstall[]
  runners: TDataModelRunner[]
  workflows: TDataModelWorkflow[]
  steps: TDataModelStep[]
  queues: TDataModelQueue[]
  emitters: TDataModelEmitter[]
  signals: TDataModelSignal[]
}

export const getDataModel = (orgID: string) =>
  api<TDataModelResponse>({ path: 'data-model', params: { org_id: orgID } })
