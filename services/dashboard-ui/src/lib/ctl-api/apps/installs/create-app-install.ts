import { api } from '@/lib/api'
import type { TInstall } from '@/types'

export type TCreateAppInstallBody = {
  aws_account?: {
    iam_role_arn: ''
    region: string
  }
  azure_account?: {
    location: string
    service_principal_app_id: ''
    service_principal_password: ''
    subscription_id: ''
    subscription_tenant_id: ''
  }
  inputs?: Record<string, string>
  install_config?: {
    approval_option: 'prompt' | 'approve-all'
  }
  metadata?: {
    managed_by: 'nuon/dashboard' // 'nuon/cli/config'
  }
  name: string
}

export const createAppInstall = ({
  appId,
  body,
  orgId,
}: {
  appId: string
  body: TCreateAppInstallBody
  orgId: string
}) =>
  api<TInstall>({
    abortTimeout: 25000,
    body,
    method: 'POST',
    orgId,
    path: `apps/${appId}/installs`,
  })
