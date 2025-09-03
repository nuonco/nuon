import type { TInstall } from '@/types'
import { mutateData } from '@/utils'

export interface ICreateInstallData {
  name: string
  inputs?: Record<string, string>
  aws_account?: {
    iam_role_arn: string
    region: string
  }
  azure_account?: {
    location: string
    service_principal_app_id?: string
    service_principal_password?: string
    subscription_id?: string
    subscription_tenant_id?: string
  }
  metadata?: {
    managed_by?: string
  }
}

export const installManagedByUI = 'nuon/dashboard'

export interface ICreateInstall {
  appId: string
  orgId: string
  data: ICreateInstallData
}

export async function createInstall({ appId, orgId, data }: ICreateInstall) {
  return mutateData<TInstall>({
    errorMessage: 'Unable to create install.',
    data: data as unknown as Record<string, unknown>,
    orgId,
    path: `apps/${appId}/installs`,
  })
}