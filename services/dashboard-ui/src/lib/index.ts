import type {
  TComponent,
  TComponentConfig,
  TInstall,
  TInstallComponent,
  TInstallDeploy,
  TOrg,
  TSandbox,
  TSandboxRun,
} from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

// CTL API fetch
// ===========================================================================
export async function getComponent({ componentId, orgId }): TComponent {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getComponentConfig({
  componentId,
  orgId,
}): TComponentConfig {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/configs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json().then((cfgs) => cfgs?.[0])
}

export async function getBuild({
  buildId,
  componentId,
  orgId,
}): Record<string, unknown> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/builds/${buildId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getInstallComponent({
  installComponentId,
  installId,
  orgId,
}): TInstallComponent {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/components`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res
    .json()
    .then((comps) => comps.find((c) => c?.id === installComponentId))
}

export async function getDeploy({
  orgId,
  installId,
  deployId,
}): TInstallDeploy {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getDeployLogs({ orgId, installId, deployId }) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getDeployPlan({ orgId, installId, deployId }) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/plan`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getSandboxRun({ orgId, installId, runId }): TSandboxRun {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/sandbox-runs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json().then((runs) => runs.find((r) => r?.id === runId))
}

export async function getSandboxRunLogs({ orgId, installId, runId }) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/sandbox-run/${runId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getInstallEvents({ installId, orgId }): TInstall {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/events`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getInstall({ installId, orgId }): TInstall {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch data')
  }

  return data.json()
}

export async function getInstalls({ orgId }): Array<TInstall> {
  const data = await fetch(`${API_URL}/v1/installs`, await getFetchOpts(orgId))

  if (!data.ok) {
    throw new Error('Failed to fetch data')
  }

  return data.json()
}

export async function getOrgs(): Array<TOrg> {
  const res = await fetch(`${API_URL}/v1/orgs`, await getFetchOpts())

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export async function getOrg({ orgId }): TOrg {
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
