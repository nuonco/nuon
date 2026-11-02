import { api } from '@/lib'
import type {
  TCloudConnection,
  TCloudConnectionSetup,
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
  platform,
  preset,
}: {
  orgId: string
  q?: string
  platform?: string
  preset?: string
} & TPaginationParams) =>
  api<TCloudConnectionSummary[]>({
    orgId,
    path: `cloud-connections${buildQueryParams({ limit, offset, q, platform, preset })}`,
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

export const getCloudConnectionSetup = ({
  connectionId,
  orgId,
  repositories = [],
  existingPrincipal = false,
}: {
  connectionId: string
  orgId: string
  repositories?: string[]
  existingPrincipal?: boolean
}) => {
  const params = new URLSearchParams()
  repositories.forEach((repository) => params.append('repository', repository))
  if (existingPrincipal) params.set('existing_principal', 'true')
  return api<TCloudConnectionSetup>({
    orgId,
    path: `cloud-connections/${connectionId}/setup${params.size ? `?${params}` : ''}`,
  })
}

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
