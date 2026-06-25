// Reserved prefix for auto-generated per-component install-level override
// inputs (Helm values / Terraform vars). Must match the Go constant
// config.ComponentOverrideInputPrefix. Synthetic input names have the shape:
//
//   nuon_component_override_v1_<kind>_<hex(componentName)>
//
// where <kind> is "helm_values" or "tf_vars" and the component name is
// hex-encoded to keep the key safe and reversible.
const COMPONENT_OVERRIDE_INPUT_PREFIX = 'nuon_component_override_v1_'
const COMPONENT_OVERRIDE_KINDS = ['helm_values', 'tf_vars'] as const

function decodeHex(encoded: string): string | null {
  if (!/^(?:[0-9a-fA-F]{2})+$/.test(encoded)) return null
  const bytes = new Uint8Array(
    encoded.match(/../g)!.map((pair) => Number.parseInt(pair, 16))
  )
  try {
    // Recover UTF-8 multibyte component names produced by Go's hex.EncodeToString.
    return new TextDecoder('utf-8', { fatal: true }).decode(bytes)
  } catch {
    return null
  }
}

export type TComponentOverrideKind = (typeof COMPONENT_OVERRIDE_KINDS)[number]

// getComponentOverrideKind returns the override axis ("helm_values" / "tf_vars")
// for a reserved synthetic input name, or null when the name is a normal input.
export function getComponentOverrideKind(
  name: string
): TComponentOverrideKind | null {
  if (!name.startsWith(COMPONENT_OVERRIDE_INPUT_PREFIX)) return null
  const rest = name.slice(COMPONENT_OVERRIDE_INPUT_PREFIX.length)
  for (const kind of COMPONENT_OVERRIDE_KINDS) {
    if (rest.startsWith(`${kind}_`)) return kind
  }
  return null
}

// getInputDisplayName maps a reserved synthetic component-override input name to
// a user-facing key like "components.<name>.helm_values". Non-override input
// names are returned unchanged. Mirrors the CLI installDiffKey helper.
//
// The name format is "<prefix><kind>_<hex(componentName)>". The kind may itself
// contain underscores (e.g. "helm_values"), but the hex-encoded component name
// never does, so the final "_<hex>" segment is always the component name and
// everything before it is the kind. Parsing the kind generically (rather than
// against a fixed list) keeps this decoder forward-compatible with new override
// kinds added on the backend without a corresponding UI change.
export function getInputDisplayName(name: string): string {
  if (!name.startsWith(COMPONENT_OVERRIDE_INPUT_PREFIX)) return name
  const rest = name.slice(COMPONENT_OVERRIDE_INPUT_PREFIX.length)
  const sep = rest.lastIndexOf('_')
  if (sep <= 0) return name
  const kind = rest.slice(0, sep)
  const component = decodeHex(rest.slice(sep + 1))
  if (component === null) return name
  return `components.${component}.${kind}`
}

// getEnabledOverrideComponent returns the component name targeted by an
// "enabled" synthetic override input, or null when name is not an enabled
// override input.
export function getEnabledOverrideComponent(name: string): string | null {
  if (!name.startsWith(COMPONENT_OVERRIDE_INPUT_PREFIX)) return null
  const rest = name.slice(COMPONENT_OVERRIDE_INPUT_PREFIX.length)
  const prefix = 'enabled_'
  if (!rest.startsWith(prefix)) return null
  return decodeHex(rest.slice(prefix.length))
}

type TEnablementComponent = {
  component_id?: string
  component_name?: string
  component_dependency_ids?: string[]
  refs?: { type?: string; name?: string }[]
}

// disabledToggleableDeps returns the names of the toggleable components in the
// transitive dependency closure of componentName that are effectively disabled.
// Dependencies are the union of declared dependencies (component_dependency_ids)
// and components whose outputs are referenced (refs of type "component"),
// mirroring the Go app.ComponentEnablementResolver. These are the actionable
// components a user must turn back on to re-enable componentName.
export function disabledToggleableDeps(
  componentName: string,
  components: TEnablementComponent[],
  effectiveEnabledByName: Record<string, boolean | undefined>
): string[] {
  const nameToId = new Map<string, string>()
  const byId = new Map<string, TEnablementComponent>()
  for (const c of components) {
    if (c.component_id) byId.set(c.component_id, c)
    if (c.component_name && c.component_id)
      nameToId.set(c.component_name, c.component_id)
  }

  const rootId = nameToId.get(componentName)
  if (!rootId) return []

  const depIds = (c: TEnablementComponent): string[] => {
    const ids = new Set<string>()
    for (const id of c.component_dependency_ids ?? []) {
      if (byId.has(id)) ids.add(id)
    }
    for (const ref of c.refs ?? []) {
      if (ref.type !== 'component' || !ref.name) continue
      const id = nameToId.get(ref.name)
      if (id) ids.add(id)
    }
    return [...ids]
  }

  const culprits = new Set<string>()
  const visited = new Set<string>([rootId])
  const queue = depIds(byId.get(rootId)!)
  while (queue.length > 0) {
    const id = queue.shift()!
    if (visited.has(id)) continue
    visited.add(id)
    const comp = byId.get(id)
    if (!comp) continue
    if (comp.component_name && effectiveEnabledByName[comp.component_name] === false) {
      culprits.add(comp.component_name)
    }
    queue.push(...depIds(comp))
  }

  return [...culprits].sort()
}

type TTitleMap = Record<string, string>

const RUNNER_STATUS_TITLES: TTitleMap = {
  active: 'Runner is healthy',
  error: 'Runner is unhealthy',
  pending: 'Runner is pending',
  provisioning: 'Runner is provisioning',
  deprovisioning: 'Runner is deprovisioning',
  deprovisioned: 'Runner is deprovisioned',
  reprovisioning: 'Runner is reprovisioning',
  offline: 'Runner is offline',
  'awaiting-install-stack-run': 'Runner is awaiting install stack run',
  unknown: 'Runner status is unknown',
}

const SANDBOX_STATUS_TITLES: TTitleMap = {
  active: 'Sandbox is provisioned',
  error: 'Sandbox has an error',
  queued: 'Sandbox is queued',
  provisioning: 'Sandbox is provisioning',
  deprovisioning: 'Sandbox is deprovisioning',
  deprovisioned: 'Sandbox is deprovisioned',
  reprovisioning: 'Sandbox is reprovisioning',
  access_error: 'Sandbox has an access error',
  deleted: 'Sandbox has been deleted',
  delete_failed: 'Sandbox deletion failed',
  empty: 'Sandbox is empty',
  unknown: 'Sandbox status is unknown',
}

const COMPONENTS_STATUS_TITLES: TTitleMap = {
  active: 'Components are deployed',
  inactive: 'Components are inactive',
  error: 'Component has an error',
  noop: 'Deployment had no changes',
  planning: 'Deployment is planning',
  syncing: 'Deployment is syncing',
  executing: 'Deployment is executing',
  cancelled: 'Deployment was cancelled',
  pending: 'Deployment is pending',
  queued: 'Deployment is queued',
  'pending-approval': 'Deployment is pending approval',
  'approval-denied': 'Deployment approval was denied',
  unknown: 'Deployment status is unknown',
}

const DEPROVISIONING_RUNNER_OVERRIDES: TTitleMap = {
  active: 'Runner waiting to teardown',
  deprovisioned: 'Runner teardown complete',
}

const DEPROVISIONED_RUNNER_OVERRIDES: TTitleMap = {
  active: 'Runner torn down',
  deprovisioned: 'Runner teardown complete',
}

const DEPROVISIONING_SANDBOX_OVERRIDES: TTitleMap = {
  active: 'Sandbox waiting to teardown',
  deprovisioned: 'Sandbox teardown complete',
  deprovisioning: 'Sandbox tearing down',
}

const DEPROVISIONED_SANDBOX_OVERRIDES: TTitleMap = {
  active: 'Sandbox torn down',
  deprovisioned: 'Sandbox teardown complete',
  deprovisioning: 'Sandbox torn down',
}

const DEPROVISIONING_COMPONENTS_OVERRIDES: TTitleMap = {
  active: 'Components waiting to teardown',
  pending: 'Components tearing down',
  executing: 'Components tearing down',
}

const DEPROVISIONED_COMPONENTS_OVERRIDES: TTitleMap = {
  active: 'Components torn down',
  pending: 'Components torn down',
  executing: 'Components torn down',
}

function getStatusTitle(
  map: TTitleMap,
  status: string,
  fallback: string
): string {
  return map[status] ?? fallback
}

export function getInstallRunnerStatusTitle(status: string): string {
  return getStatusTitle(
    RUNNER_STATUS_TITLES,
    status,
    RUNNER_STATUS_TITLES.unknown
  )
}

export function getInstallSandboxStatusTitle(status: string): string {
  return getStatusTitle(
    SANDBOX_STATUS_TITLES,
    status,
    SANDBOX_STATUS_TITLES.unknown
  )
}

export function getInstallComponentsStatusTitle(status: string): string {
  return getStatusTitle(
    COMPONENTS_STATUS_TITLES,
    status,
    COMPONENTS_STATUS_TITLES.unknown
  )
}

export function getInstallStatusTitle(
  statusKey: string,
  status: string,
  lifecycleStatus?: string
): string {
  if (lifecycleStatus === 'deprovisioning' || lifecycleStatus === 'deprovisioned') {
    const isFinished = lifecycleStatus === 'deprovisioned'
    let override: string | undefined
    switch (statusKey) {
      case 'runner_status':
        override = (isFinished ? DEPROVISIONED_RUNNER_OVERRIDES : DEPROVISIONING_RUNNER_OVERRIDES)[status]
        break
      case 'sandbox_status':
        override = (isFinished ? DEPROVISIONED_SANDBOX_OVERRIDES : DEPROVISIONING_SANDBOX_OVERRIDES)[status]
        break
      case 'composite_component_status':
        override = (isFinished ? DEPROVISIONED_COMPONENTS_OVERRIDES : DEPROVISIONING_COMPONENTS_OVERRIDES)[status]
        break
    }
    if (override) return override
  }

  switch (statusKey) {
    case 'runner_status':
      return getInstallRunnerStatusTitle(status)
    case 'sandbox_status':
      return getInstallSandboxStatusTitle(status)
    case 'composite_component_status':
      return getInstallComponentsStatusTitle(status)
    default:
      return 'Waiting on status'
  }
}
