import { api } from '@/lib/api'
import type { TAWSAccountConnection } from '@/types'

export const createAWSAccountConnection = ({
  body,
  orgId,
}: {
  body: { name: string; account_id: string; default_region: string }
  orgId: string
}) =>
  api<TAWSAccountConnection>({
    method: 'POST',
    orgId,
    path: 'aws-account-connections',
    body,
  })

export const getAWSAccountConnections = ({ orgId }: { orgId: string }) =>
  api<TAWSAccountConnection[]>({ orgId, path: 'aws-account-connections' })

export const getAWSAccountConnection = ({
  connectionId,
  orgId,
}: {
  connectionId: string
  orgId: string
}) =>
  api<TAWSAccountConnection>({
    orgId,
    path: `aws-account-connections/${connectionId}`,
  })

export const updateAWSAccountConnection = ({
  connectionId,
  body,
  orgId,
}: {
  connectionId: string
  body: { name?: string; default_region?: string; role_arn?: string }
  orgId: string
}) =>
  api<TAWSAccountConnection>({
    method: 'PATCH',
    orgId,
    path: `aws-account-connections/${connectionId}`,
    body,
  })

export const verifyAWSAccountConnection = ({
  connectionId,
  orgId,
}: {
  connectionId: string
  orgId: string
}) =>
  api<TAWSAccountConnection>({
    abortTimeout: 30000,
    method: 'POST',
    orgId,
    path: `aws-account-connections/${connectionId}/verify`,
  })
