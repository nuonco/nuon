export type TFleetLifecycle =
  | 'provisioning'
  | 'provisioned'
  | 'deprovisioning'
  | 'deprovisioned'
  | 'reprovisioning'

export type TFleetHealth =
  | 'healthy'
  | 'degraded'
  | 'unhealthy'
  | 'progressing'
  | 'unknown'

export type TFleetResource = {
  id: string
  name: string
  kind: string
  health?: TFleetHealth
}

export type TFleetInstall = {
  id: string
  name: string
  labels: Record<string, string>
  branchName?: string
  lifecycle: TFleetLifecycle
  hasRunner: boolean
  runnerStatus?: string
  health?: TFleetHealth
  componentStatus?: string
  commitsBehind?: number
  configDriftCount: number
  resources?: TFleetResource[]
}

export type TFleetRunnerFilter = 'active' | 'inactive'

export type TFleetFilters = {
  labels: Set<string>
  runner?: TFleetRunnerFilter
  needsUpdate: boolean
}

const COLUMN = 288
const COLUMNS_PER_ROW = 4
const CARD_WIDTH = 256
const CARD_HEIGHT = 208
const HEADER_HEIGHT = 24
const CARD_STEP = 232

export type TFleetSection = {
  id: string
  title: string
  installs: TFleetInstall[]
}

export type TFleetLayoutSection = {
  id: string
  title: string
  count: number
  x: number
  y: number
  width: number
  installs: { install: TFleetInstall; x: number; y: number }[]
}

export const runnerLabel = (install: TFleetInstall) =>
  install.hasRunner ? install.runnerStatus || 'unknown' : 'No runner'

export const trackingLabel = (install: TFleetInstall) => {
  if (install.commitsBehind && install.commitsBehind > 0) {
    return install.commitsBehind === 1
      ? '1 commit behind'
      : `${install.commitsBehind} commits behind`
  }
  if (install.configDriftCount > 0) {
    return install.configDriftCount === 1
      ? '1 drifted'
      : `${install.configDriftCount} drifted`
  }
  if (install.commitsBehind === 0) return 'Up to date'
  return 'Not reported'
}

export const installResources = (install: TFleetInstall): TFleetResource[] => {
  if (install.resources?.length) return install.resources
  const health = install.health
  return [
    { id: `${install.id}-stack`, name: 'stack', kind: 'Stack', health },
    { id: `${install.id}-sandbox`, name: 'sandbox', kind: 'Sandbox', health },
    { id: `${install.id}-workload`, name: 'api', kind: 'Deployment', health },
  ]
}

export const layoutInstallResources = (resources: TFleetResource[]) =>
  resources.map((resource, index) => ({
    resource,
    x: 0,
    y: index * 112,
  }))

export const labelPairs = (install: TFleetInstall) =>
  Object.entries(install.labels)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => `${key}:${value}`)

export const labelOptions = (installs: TFleetInstall[]) => {
  const map = new Map<string, Set<string>>()
  for (const install of installs) {
    for (const [key, value] of Object.entries(install.labels)) {
      const values = map.get(key) ?? new Set<string>()
      values.add(value)
      map.set(key, values)
    }
  }
  return [...map.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .flatMap(([key, values]) =>
      [...values].sort().map((value) => `${key}:${value}`)
    )
}

export const passesFilters = (
  install: TFleetInstall,
  filters: TFleetFilters
) => {
  if (filters.labels.size) {
    const have = new Set(labelPairs(install))
    if (![...filters.labels].every((label) => have.has(label))) return false
  }
  const active = install.hasRunner && install.runnerStatus === 'active'
  if (filters.runner === 'active' && !active) return false
  if (filters.runner === 'inactive' && active) return false
  if (
    filters.needsUpdate &&
    !(install.commitsBehind && install.commitsBehind > 0)
  )
    return false
  return true
}

export const searchHaystack = (install: TFleetInstall) =>
  [
    install.name,
    install.branchName,
    install.lifecycle,
    runnerLabel(install),
    install.health,
    install.componentStatus,
    trackingLabel(install),
    ...labelPairs(install),
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()

export const matchesSearch = (install: TFleetInstall, query: string) => {
  const needle = query.trim().toLowerCase()
  if (!needle) return false
  return searchHaystack(install).includes(needle)
}

export const groupFleetInstalls = (
  installs: TFleetInstall[],
  currentBranch?: string
): TFleetSection[] => {
  const byBranch = new Map<string, TFleetInstall[]>()
  for (const install of installs) {
    const key = install.branchName || ''
    byBranch.set(key, [...(byBranch.get(key) ?? []), install])
  }
  for (const list of byBranch.values()) {
    list.sort(
      (a, b) => a.name.localeCompare(b.name) || a.id.localeCompare(b.id)
    )
  }
  const rank = (key: string) => {
    if (!key) return 2
    if (currentBranch && key === currentBranch) return 0
    return 1
  }
  return [...byBranch.keys()]
    .sort((a, b) => rank(a) - rank(b) || a.localeCompare(b))
    .map((key) => ({
      id: key || 'not-tracking',
      title: key || 'Not tracking',
      installs: byBranch.get(key) ?? [],
    }))
}

export const layoutFleet = (
  sections: TFleetSection[]
): TFleetLayoutSection[] => {
  const placed: TFleetLayoutSection[] = []
  let rowTop = 0
  for (let start = 0; start < sections.length; start += COLUMNS_PER_ROW) {
    const row = sections.slice(start, start + COLUMNS_PER_ROW)
    let rowBottom = rowTop + HEADER_HEIGHT
    row.forEach((section, column) => {
      const x = column * COLUMN
      const cards = section.installs.map((install, index) => ({
        install,
        x,
        y: rowTop + HEADER_HEIGHT + 16 + index * CARD_STEP,
      }))
      const bottom = cards.length
        ? cards[cards.length - 1].y + CARD_HEIGHT
        : rowTop + HEADER_HEIGHT
      rowBottom = Math.max(rowBottom, bottom)
      placed.push({
        id: section.id,
        title: section.title,
        count: section.installs.length,
        x,
        y: rowTop,
        width: CARD_WIDTH,
        installs: cards,
      })
    })
    rowTop = rowBottom + 40
  }
  return placed
}
