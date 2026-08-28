import { api } from '@/lib/api'

export type TCreateCustomerManagedInstallBody = {
  app_id: string
  intended_name: string
  release_id: string
  telemetry: string
  aws_region: string
  aws_account_id?: string
  inputs: Record<string, string | null>
}

export type TCustomerManagedInstall = {
  install: { id: string; name: string }
  portal_service_account: { id: string; name: string; email: string }
}

export const createCustomerManagedInstall = ({
  body,
  orgId,
}: {
  body: TCreateCustomerManagedInstallBody
  orgId: string
}) =>
  api<TCustomerManagedInstall>({
    body,
    method: 'POST',
    orgId,
    path: 'customer-managed/installs',
  })
