import { TInstall, TInstallComponent, TSandboxRun } from '@/types'

// general utils
export const API_URL =
  process?.env?.NEXT_PUBLIC_API_URL || 'https://api.nuon.co'
export const POLL_DURATION =
  (process?.env?.NEXT_PUBLIC_POLL_DURATION as unknown as number) || 45000
export const SHORT_POLL_DURATION = (process?.env?.NEXT_PUBLIC_SHORT_POLL_DURATION as unknown as number) || 22500
export const GITHUB_APP_NAME = process?.env?.NEXT_PUBLIC_GITHUB_APP_NAME || "nat-test-local"

export const sentanceCase = (s = '') => s.charAt(0).toUpperCase() + s.slice(1)
export const titleCase = (s = '') =>
  s.replace(/^_*(.)|_+(.)/g, (s, c, d) => (c ? c.toUpperCase() : ' ' + d))

export function getFlagEmoji(countryCode = 'us') {
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
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
  const sandboxStatus = getSandboxStatus(install?.install_sandbox_runs as Array<TSandboxRun> || [])
  const componentStatus = getInstallComponentStatus(
    install?.install_components as Array<TInstallComponent> || []
  )

  return {
    componentStatus,
    installStatus: getInstallStatus([sandboxStatus, componentStatus]),
    sandboxStatus,
  }
}

export * from './install-regions'
export * from './get-fetch-opts'
export * from './datadog-logs'
export * from './datadog-rum'
