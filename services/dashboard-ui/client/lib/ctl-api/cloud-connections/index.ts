import { api } from '@/lib/api'
import type { TCloudConnection } from '@/types'

export const createCloudConnection = ({
  body,
  orgId,
}: {
  body: {
    name: string
    platform: 'aws'
    target_id: string
    principal: string
    default_region?: string
    preset: 'stacks' | 'custom'
  }
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
    abortTimeout: 30000,
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
