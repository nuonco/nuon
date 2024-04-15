import type {
  TComponent,
  TComponentConfig,
  TInstall,
  TInstallComponent,
  TInstallDeploy,
  TInstallEvent,
  TOrg,
  TSandbox,
  TSandboxRun,
} from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

// CTL API fetch
// ===========================================================================

interface IGetComponent {
  componentId: string
  orgId: string
}

export async function getComponent({
  componentId,
  orgId,
}: IGetComponent): Promise<TComponent> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetComponentConfig {
  componentId: string
  orgId: string
}

export async function getComponentConfig({
  componentId,
  orgId,
}: IGetComponentConfig): Promise<TComponentConfig> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/configs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json().then((cfgs) => cfgs?.[0])
}

interface IGetBuild {
  buildId: string
  componentId: string
  orgId: string
}

export async function getBuild({
  buildId,
  componentId,
  orgId,
}: IGetBuild): Promise<Record<string, unknown>> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/builds/${buildId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetInstallComponent {
  installComponentId: string
  installId: string
  orgId: string
}

export async function getInstallComponent({
  installComponentId,
  installId,
  orgId,
}: IGetInstallComponent): Promise<TInstallComponent> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/components`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res
    .json()
    .then((comps: Array<TInstallComponent>) =>
      comps.find((c) => c?.id === installComponentId)
    )
}

interface IGetDeploy {
  deployId: string
  installId: string
  orgId: string
}

export async function getDeploy({
  orgId,
  installId,
  deployId,
}: IGetDeploy): Promise<TInstallDeploy> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetDeployLogs {
  deployId: string
  installId: string
  orgId: string
}

export async function getDeployLogs({
  orgId,
  installId,
  deployId,
}: IGetDeployLogs) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetDeployPlan {
  deployId: string
  installId: string
  orgId: string
}

export async function getDeployPlan({
  orgId,
  installId,
  deployId,
}: IGetDeployPlan) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/plan`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetSandboxRun {
  installId: string
  orgId: string
  runId: string
}

export async function getSandboxRun({
  orgId,
  installId,
  runId,
}: IGetSandboxRun): Promise<TSandboxRun> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/sandbox-runs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res
    .json()
    .then((runs: Array<TSandboxRun>) => runs.find((r) => r?.id === runId))
}

interface IGetSandboxRunLogs {
  installId: string
  orgId: string
  runId: string
}

export async function getSandboxRunLogs({
  orgId,
  installId,
  runId,
}: IGetSandboxRunLogs) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/sandbox-run/${runId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetInstallEvents {
  installId: string
  orgId: string
}

export async function getInstallEvents({
  installId,
  orgId,
}: IGetInstallEvents): Promise<Array<TInstallEvent>> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/events`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetInstall {
  installId: string
  orgId: string
}

export async function getInstall({
  installId,
  orgId,
}: IGetInstall): Promise<TInstall> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch data')
  }

  return data.json()
}

interface IGetInstalls {
  orgId: string
}

export async function getInstalls({
  orgId,
}: IGetInstalls): Promise<Array<TInstall>> {
  const data = await fetch(`${API_URL}/v1/installs`, await getFetchOpts(orgId))

  if (!data.ok) {
    throw new Error('Failed to fetch data')
  }

  return data.json()
}

export async function getOrgs(): Promise<Array<TOrg>> {
  const res = await fetch(`${API_URL}/v1/orgs`, await getFetchOpts())

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

interface IGetOrg {
  orgId: string
}

export async function getOrg({ orgId }: IGetOrg): Promise<TOrg> {
  const data = await fetch(
    `${API_URL}/v1/orgs/current`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch data')
  }

  return data.json()
}

// Local API fetch
// ==========================================================================
