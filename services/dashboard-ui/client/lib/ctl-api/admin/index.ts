import { api } from '@/lib/api'
import type { TRunner } from '@/types'

type AdminBase = {
  adminApiUrl: string
}

type AdminMutation = AdminBase & {
  adminEmail: string
}

const adminHeaders = (adminEmail: string) => ({ 'X-Nuon-Admin-Email': adminEmail })

export const adminGetOrgRunner = ({ orgId, adminApiUrl }: { orgId: string } & AdminBase) =>
  api<TRunner>({ baseUrl: adminApiUrl, path: `orgs/${orgId}/admin-get-runner` })

export const adminGetInstallRunner = ({ installId, adminApiUrl }: { installId: string } & AdminBase) =>
  api<TRunner>({ baseUrl: adminApiUrl, path: `installs/${installId}/admin-get-runner` })

export const adminGetOrgFeaturesList = ({ adminApiUrl }: AdminBase) =>
  api<string[]>({ baseUrl: adminApiUrl, path: `orgs/admin-features` })

export const adminAddSupportUsersToOrg = ({ orgId, adminApiUrl, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-support-users` })

export const adminRemoveSupportUsersFromOrg = ({ orgId, adminApiUrl, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-remove-support-users` })

export const adminReprovisionOrg = ({ orgId, adminApiUrl, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-reprovision` })

export const adminRestartOrg = ({ orgId, adminApiUrl, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-restart` })

export const adminRestartOrgRunners = ({ orgId, adminApiUrl, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-restart-runners` })

export const adminEnableOrgDebugMode = ({ orgId, adminApiUrl, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-debug-mode` })

export const adminUpdateOrgFeatures = ({ orgId, features, adminApiUrl, adminEmail }: { orgId: string; features: Record<string, boolean> } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'PATCH', body: { features }, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-features` })

export const adminReprovisionApp = ({ appId, adminApiUrl, adminEmail }: { appId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `apps/${appId}/admin-reprovision` })

export const adminRestartApp = ({ appId, adminApiUrl, adminEmail }: { appId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `apps/${appId}/admin-restart` })

export const adminReprovisionInstall = ({ installId, adminApiUrl, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-reprovision` })

export const adminReprovisionInstallRunner = ({ runnerId, adminApiUrl, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/reprovision` })

export const adminRestartInstall = ({ installId, adminApiUrl, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-restart` })

export const adminTeardownInstallComponents = ({ installId, orgId }: { installId: string; orgId: string }) =>
  api<void>({ method: 'POST', body: {}, orgId, path: `installs/${installId}/components/teardown-all` })

export const adminUpdateInstallSandbox = ({ installId, adminApiUrl, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-update-sandbox` })

export const adminRestartRunner = ({ runnerId, adminApiUrl, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/restart` })

export const adminGracefulRunnerShutdown = ({ runnerId, adminApiUrl, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/graceful-shutdown` })

export const adminForceRunnerShutdown = ({ runnerId, adminApiUrl, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/force-shutdown` })

export const adminInvalidateRunnerToken = ({ runnerId, adminApiUrl, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/invalidate-service-account-token` })

export const adminShutdownRunnerJob = ({ installId, adminApiUrl, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/runners/shutdown-job` })
