import { api } from '@/lib/api'
import type {
  TAdminSandboxConfig,
  TSandboxRunner,
  TSandboxTemplates,
  TSandboxJob,
} from '@/types'

type AdminMutation = {
  adminEmail: string
}

const adminHeaders = (adminEmail: string) => ({ 'X-Nuon-Admin-Email': adminEmail })

export const adminListSandboxRunners = () =>
  api<TSandboxRunner[]>({ baseUrl: '/admin', path: `runners/sandbox` })

export const adminGetSandboxTemplates = () =>
  api<TSandboxTemplates>({ baseUrl: '/admin', path: `runners/sandbox/templates` })

export const adminGetSandboxConfigs = ({ runnerId }: { runnerId: string }) =>
  api<TAdminSandboxConfig[]>({ baseUrl: '/admin', path: `runners/${runnerId}/sandbox-configs` })

export const adminUpsertSandboxConfig = ({
  runnerId,
  adminEmail,
  body,
}: { runnerId: string; body: Partial<TAdminSandboxConfig> } & AdminMutation) =>
  api<TAdminSandboxConfig>({
    baseUrl: '/admin',
    method: 'PUT',
    headers: adminHeaders(adminEmail),
    path: `runners/${runnerId}/sandbox-configs`,
    body,
  })

export const adminDeleteSandboxConfig = ({
  runnerId,
  configId,
  adminEmail,
}: { runnerId: string; configId: string } & AdminMutation) =>
  api<void>({
    baseUrl: '/admin',
    method: 'DELETE',
    headers: adminHeaders(adminEmail),
    path: `runners/${runnerId}/sandbox-configs/${configId}`,
  })

export const adminResetSandboxConfigs = ({
  runnerId,
  adminEmail,
}: { runnerId: string } & AdminMutation) =>
  api<void>({
    baseUrl: '/admin',
    method: 'POST',
    body: {},
    headers: adminHeaders(adminEmail),
    path: `runners/${runnerId}/sandbox-configs/reset`,
  })

export const adminGetSandboxJobs = ({ runnerId }: { runnerId: string }) =>
  api<TSandboxJob[]>({ baseUrl: '/admin', path: `runners/${runnerId}/sandbox-jobs` })
