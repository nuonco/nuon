import { api } from '@/lib/api'

export type TInstallationRegistration = {
  schema_version: number
  registration_id: string
  release_id: string
  release_digest: string
  package_id: string
  package_digest: string
  bundle_digest: string
  archive_digest: string
  deployment_id: string
  install_id: string
  cloud: { provider: string; account_id: string; region: string }
  stack: { type: string; id: string; name: string }
  installed_at: string
}

export type TInstallRegistrationResponse = {
  install: { id: string; name: string; app_id: string; app_config_id: string }
  registration: { id: string; release_id: string; package_id: string }
  management_policy: {
    connectivity: string
    release_selection: string
    command_authority: string
    approval_authority: string
    telemetry: string
  }
  release_deployment: { id: string; release_id: string; package_id: string }
}

export const registerCustomerManagedInstall = ({
  body,
  orgId,
}: {
  body: TInstallationRegistration
  orgId: string
}) =>
  api<TInstallRegistrationResponse>({
    body,
    method: 'POST',
    orgId,
    path: 'install-registrations',
  })
