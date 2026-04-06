import { api } from '@/lib/api'
import type { TRunner } from '@/types'

type AdminMutation = {
  adminEmail: string
}

const adminHeaders = (adminEmail: string) => ({ 'X-Nuon-Admin-Email': adminEmail })

export const adminGetOrgRunner = ({ orgId }: { orgId: string }) =>
  api<TRunner>({ baseUrl: '/admin', path: `orgs/${orgId}/admin-get-runner` })

export const adminGetInstallRunner = ({ installId }: { installId: string }) =>
  api<TRunner>({ baseUrl: '/admin', path: `installs/${installId}/admin-get-runner` })

export const adminGetOrgFeaturesList = () =>
  api<string[]>({ baseUrl: '/admin', path: `orgs/admin-features` })

export const adminAddSupportUsersToOrg = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-support-users` })

export const adminRemoveSupportUsersFromOrg = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-remove-support-users` })

export const adminReprovisionOrg = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-reprovision` })

export const adminRestartOrg = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-restart` })

export const adminRestartOrgRunners = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-restart-runners` })

export const adminEnableOrgDebugMode = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-debug-mode` })

export const adminUpdateOrgFeatures = ({ orgId, features, adminEmail }: { orgId: string; features: Record<string, boolean> } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'PATCH', body: { features }, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-features` })

export const adminReprovisionApp = ({ appId, adminEmail }: { appId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `apps/${appId}/admin-reprovision` })

export const adminRestartApp = ({ appId, adminEmail }: { appId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `apps/${appId}/admin-restart` })

export const adminReprovisionInstall = ({ installId, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-reprovision` })

export const adminReprovisionInstallRunner = ({ runnerId, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/reprovision` })

export const adminRestartInstall = ({ installId, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-restart` })

export const adminRestartInstallQueues = ({ installId, adminApiUrl, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: adminApiUrl, method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-restart-queues` })

export const adminTeardownInstallComponents = ({ installId, orgId }: { installId: string; orgId: string }) =>
  api<void>({ method: 'POST', body: {}, orgId, path: `installs/${installId}/components/teardown-all` })

export const adminUpdateInstallSandbox = ({ installId, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/admin-update-sandbox` })

export const adminRestartRunner = ({ runnerId, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/restart` })

export const adminGracefulRunnerShutdown = ({ runnerId, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/graceful-shutdown` })

export const adminForceRunnerShutdown = ({ runnerId, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/force-shutdown` })

export const adminInvalidateRunnerToken = ({ runnerId, adminEmail }: { runnerId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `runners/${runnerId}/invalidate-service-account-token` })

export const adminShutdownRunnerJob = ({ installId, adminEmail }: { installId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `installs/${installId}/runners/shutdown-job` })

export const adminDeprovisionOrg = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-deprovision` })

export const adminForgetOrgInstalls = ({ orgId, adminEmail }: { orgId: string } & AdminMutation) =>
  api<void>({ baseUrl: '/admin', method: 'POST', body: {}, headers: adminHeaders(adminEmail), path: `orgs/${orgId}/admin-forget-installs` })
