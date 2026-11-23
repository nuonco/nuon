import { api } from '@/lib/api'
import type {
  TAppFleet,
  TAppFleetFilters,
  TAppFleetInstall,
  TPaginationParams,
} from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const FLEET_NONE = '__none__'

export interface IGetAppFleet extends TAppFleetFilters {
  appId: string
  orgId: string
  per_group?: number
}

export const getAppFleet = ({ appId, orgId, ...params }: IGetAppFleet) =>
  api<TAppFleet>({
    orgId,
    path: `apps/${appId}/fleet${buildQueryParams(params)}`,
  })

export interface IGetAppFleetInstalls
  extends TAppFleetFilters,
    TPaginationParams {
  appId: string
  orgId: string
  app_branch_id?: string
  app_branch_group?: string
}

export const getAppFleetInstalls = ({
  appId,
  orgId,
  ...params
}: IGetAppFleetInstalls) =>
  api<TAppFleetInstall[]>({
    orgId,
    path: `apps/${appId}/fleet/installs${buildQueryParams(params)}`,
    paginated: true,
  })
