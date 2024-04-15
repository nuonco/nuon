import { getSession } from '@auth0/nextjs-auth0'
import { TInstall, TInstallComponent, TSandboxRun } from "@/types"

// general utils
export const API_URL = process?.env?.NEXT_PUBLIC_API_URL

export const sentanceCase = (s = '') => s.charAt(0).toUpperCase() + s.slice(1)
export const titleCase = (s = '') =>
  s.replace(/^_*(.)|_+(.)/g, (s, c, d) => (c ? c.toUpperCase() : ' ' + d))

export function getFlagEmoji(countryCode = "us") {
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

// fetch helper
export async function getFetchOpts(orgId = ""): Promise<RequestInit> {
  const session = await getSession()
  return {
    cache: 'no-store',
    headers: {
      Authorization: `Bearer ${session?.accessToken}`,
      'Content-Type': 'application/json',
      'X-Nuon-Org-ID': orgId,
    },
  }
}

// install status helpers
type TStatus = { status: string; status_description: string }

export function getSandboxStatus(runs: TSandboxRun[]): TStatus {
  return {
    status: runs?.[0]?.status || 'error',
    status_description:
      runs?.[0]?.status_description || "Sandbox isn't provisioned",
  }
}

export function getInstallComponentStatus(
  components: TInstallComponent[]
): TStatus {
  let status = {
    status: 'waiting',
    status_description: 'Waiting on components to deploy',
  }

  if (
    components.some(
      (c) =>
        c?.install_deploys?.[0]?.status === 'failed' ||
        c?.install_deploys?.[0]?.status === 'error'
    )
  ) {
    status = {
      status: 'failed',
      status_description: 'Some components have failed to deploy',
    }
  }

  if (components.every((c) => c?.install_deploys?.[0]?.status === 'active')) {
    status = {
      status: 'active',
      status_description: 'All components are active',
    }
  }

  if (components?.every((c) => c?.install_deploys?.length === 0)) {
    status = {
      status: 'waiting',
      status_description: 'Components are not deployed',
    }
  }

  return status
}

export function getInstallStatus(statuses: TStatus[]): TStatus {
  let status: TStatus = {
    status: 'waiting',
    status_description: 'Install is waiting for something',
  }

  if (statuses.some((s) => s?.status === 'failed' || s?.status === 'error')) {
    status = {
      status: 'error',
      status_description: 'Something has gone wrong',
    }
  }

  if (statuses.every((s) => s?.status === 'active')) {
    status = {
      status: 'active',
      status_description: 'Everything is working',
    }
  }

  return status
}

export type TFullInstallStatus = Record<
  'componentStatus' | 'installStatus' | 'sandboxStatus',
  TStatus
>

export function getFullInstallStatus(install: TInstall): TFullInstallStatus {
  const sandboxStatus = getSandboxStatus(install?.install_sandbox_runs || [])
  const componentStatus = getInstallComponentStatus(
    install?.install_components || []
  )

  return {
    componentStatus,
    installStatus: getInstallStatus([sandboxStatus, componentStatus]),
    sandboxStatus,
  }
}

export * from '@/utils/install-regions'
