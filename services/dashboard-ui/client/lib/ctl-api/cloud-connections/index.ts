import { api } from '@/lib'
import type {
  TCloudConnection,
  TCloudConnectionSummary,
  TCreateCloudConnectionRequest,
  TPaginationParams,
} from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const createCloudConnection = ({
  body,
  orgId,
}: {
  body: TCreateCloudConnectionRequest
  orgId: string
}) =>
  api<TCloudConnection>({
    method: 'POST',
    orgId,
    path: 'cloud-connections',
    body,
  })

export const getCloudConnections = ({
  orgId,
  limit,
  offset,
  q,
}: {
  orgId: string
  q?: string
} & TPaginationParams) =>
  api<TCloudConnectionSummary[]>({
    orgId,
    path: `cloud-connections${buildQueryParams({ limit, offset, q })}`,
    paginated: true,
  })

export const getCloudConnection = ({
  connectionId,
  orgId,
}: {
  connectionId: string
  orgId: string
}) =>
  api<TCloudConnection>({
    orgId,
    path: `cloud-connections/${connectionId}`,
  })

export const verifyCloudConnection = ({
  connectionId,
  orgId,
}: {
  connectionId: string
  orgId: string
}) =>
  api<TCloudConnection>({
    method: 'POST',
    orgId,
    path: `cloud-connections/${connectionId}/verify`,
  })

export const deleteCloudConnection = ({
  connectionId,
  orgId,
}: {
  connectionId: string
  orgId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `cloud-connections/${connectionId}`,
  })
