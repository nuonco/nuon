import { api } from '@/lib'
import type {
  TCloudConnection,
  TCloudConnectionSetup,
  TCreateCloudConnectionRequest,
} from '@/types'

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

export const getCloudConnections = ({ orgId }: { orgId: string }) =>
  api<TCloudConnection[]>({ orgId, path: 'cloud-connections' })

export const getCloudConnectionSetup = ({
  connectionId,
  orgId,
}: {
  connectionId: string
  orgId: string
}) =>
  api<TCloudConnectionSetup>({
    orgId,
    path: `cloud-connections/${connectionId}/setup`,
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
    abortTimeout: 90000,
    method: 'POST',
    orgId,
    path: `cloud-connections/${connectionId}/verify`,
    body: {},
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
