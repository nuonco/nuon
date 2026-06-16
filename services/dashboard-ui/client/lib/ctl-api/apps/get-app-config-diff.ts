import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'

export type TDiffOp = 'add' | 'remove' | 'change' | 'noop' | ''

export type TDiffKey = {
  op: TDiffOp
  diff: string
}

export type TDiffNode = {
  key: string
  diff?: TDiffKey
  children?: TDiffNode[]
}

export type TDiffSummary = {
  has_changed: boolean
  added: number
  removed: number
  changed: number
  unchanged: number
}

export type TAppConfigDiffResponse = {
  config_id: string
  old_config_id?: string
  diff: TDiffNode
  summary: TDiffSummary
  changed: string
}

export const getAppConfigDiff = ({
  appId,
  configId,
  oldConfigId,
  orgId,
}: {
  appId: string
  configId: string
  oldConfigId?: string
  orgId: string
}) =>
  api<TAppConfigDiffResponse>({
    path: `apps/${appId}/configs/${configId}/diff${buildQueryParams({ old_config_id: oldConfigId })}`,
    orgId,
  })
