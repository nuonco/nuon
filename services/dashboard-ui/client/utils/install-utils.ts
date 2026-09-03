// Reserved prefix for auto-generated per-component install-level override
// inputs (Helm values, Terraform vars, enabled toggle, ...). Must match the Go
// constant config.ComponentOverrideInputPrefix. Synthetic input names have the
// shape:
//
//   nuon_component_override_v1_<kind>_<hex(componentName)>
//
// where <kind> is an override axis (e.g. "helm_values", "tf_vars", "enabled")
// and the component name is hex-encoded to keep the key safe and reversible.
const COMPONENT_OVERRIDE_INPUT_PREFIX = 'nuon_component_override_v1_'

// Reserved input group that holds all synthetic per-component override inputs.
// Must match the Go constant config.ComponentOverrideInputGroup.
export const COMPONENT_OVERRIDE_INPUT_GROUP = 'nuon_component_overrides'

// Override kinds whose value is structured config (rendered in a syntax-
// highlighted code block). Other kinds (e.g. "enabled") render as plain text.
// This is the only place that needs updating when a new *code* override kind is
// added; display-name decoding below is generic over kind.
const STRUCTURED_OVERRIDE_KINDS = ['helm_values', 'tf_vars'] as const

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

export type TComponentOverrideKind = (typeof STRUCTURED_OVERRIDE_KINDS)[number]

type TParsedComponentOverride = { kind: string; component: string }

// parseComponentOverrideInput decodes a reserved synthetic component-override
// input name (nuon_component_override_v1_<kind>_<hex(componentName)>) into its
// kind and component name. It is generic over kind — the component name is the
// trailing hex segment and the kind is everything before it — so new override
// axes (e.g. "enabled") never fall through to the raw id. Returns null for
// non-override names or malformed/undecodable ones. Mirrors the Go
// config.ParseComponentOverrideInputName helper.
export function parseComponentOverrideInput(
  name: string
): TParsedComponentOverride | null {
  if (!name.startsWith(COMPONENT_OVERRIDE_INPUT_PREFIX)) return null
  const rest = name.slice(COMPONENT_OVERRIDE_INPUT_PREFIX.length)
  const sep = rest.lastIndexOf('_')
  if (sep <= 0) return null
  const kind = rest.slice(0, sep)
  const component = decodeHex(rest.slice(sep + 1))
  if (component === null) return null
  return { kind, component }
}

// getComponentOverrideKind returns the structured override axis ("helm_values" /
// "tf_vars") for a reserved synthetic input name, or null when the name is a
// normal input or a non-structured override (e.g. "enabled"). Used to decide
// whether a value renders as a code block.
export function getComponentOverrideKind(
  name: string
): TComponentOverrideKind | null {
  const parsed = parseComponentOverrideInput(name)
  if (!parsed) return null
  return (STRUCTURED_OVERRIDE_KINDS as readonly string[]).includes(parsed.kind)
    ? (parsed.kind as TComponentOverrideKind)
    : null
}

// getInputDisplayName maps a reserved synthetic component-override input name to
// a user-facing key like "components.<name>.<kind>" (e.g.
// "components.api_gateway.enabled"). Non-override input names are returned
// unchanged. Mirrors the CLI installDiffKey helper.
export function getInputDisplayName(name: string): string {
  const parsed = parseComponentOverrideInput(name)
  if (!parsed) return name
  return `components.${parsed.component}.${parsed.kind}`
}

export type TComponentType = 'terraform_module' | 'helm_chart'

// A single toggleable/configurable component reconstructed from the flat
// synthetic override inputs the API exposes. `enabledInput` drives the on/off
// toggle (present iff the component is toggleable); `configInput` drives the
// tf_vars/helm_values editor (present iff the component takes structured config).
export type TComponentOverrideCard<I> = {
  component: string
  componentType?: TComponentType
  enabledInput: I | null
  configInput: I | null
  configKind: TComponentOverrideKind | null
  index: number
}

// groupComponentOverrideInputs regroups the flat per-component override inputs
// (which the API/config expose as separate inputs) into one entry per
// component, keyed off the component name encoded in each input's reserved name.
// Non-override inputs are ignored. The result is ordered by each component's
// earliest input index so display order stays stable across syncs.
export function groupComponentOverrideInputs<
  I extends { name?: string; index?: number },
>(inputs: I[]): TComponentOverrideCard<I>[] {
  const byComponent = new Map<string, TComponentOverrideCard<I>>()

  for (const input of inputs) {
    const parsed = input.name ? parseComponentOverrideInput(input.name) : null
    if (!parsed) continue

    let card = byComponent.get(parsed.component)
    if (!card) {
      card = {
        component: parsed.component,
        enabledInput: null,
        configInput: null,
        configKind: null,
        index: input.index ?? 0,
      }
      byComponent.set(parsed.component, card)
    }
    card.index = Math.min(card.index, input.index ?? 0)

    if (parsed.kind === 'enabled') {
      card.enabledInput = input
    } else if (parsed.kind === 'tf_vars' || parsed.kind === 'helm_values') {
      card.configInput = input
      card.configKind = parsed.kind
      card.componentType =
        parsed.kind === 'tf_vars' ? 'terraform_module' : 'helm_chart'
    }
  }

  return [...byComponent.values()].sort((a, b) => a.index - b.index)
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
  disabled: 'Runner is disabled',
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

export function isCustomerManagedInstall(
  install:
    | { operating_model?: { approval_authority?: string } }
    | undefined
    | null
): boolean {
  return install?.operating_model?.approval_authority === 'customer'
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
  if (
    lifecycleStatus === 'deprovisioning' ||
    lifecycleStatus === 'deprovisioned'
  ) {
    const isFinished = lifecycleStatus === 'deprovisioned'
    let override: string | undefined
    switch (statusKey) {
      case 'runner_status':
        override = (
          isFinished
            ? DEPROVISIONED_RUNNER_OVERRIDES
            : DEPROVISIONING_RUNNER_OVERRIDES
        )[status]
        break
      case 'sandbox_status':
        override = (
          isFinished
            ? DEPROVISIONED_SANDBOX_OVERRIDES
            : DEPROVISIONING_SANDBOX_OVERRIDES
        )[status]
        break
      case 'composite_component_status':
        override = (
          isFinished
            ? DEPROVISIONED_COMPONENTS_OVERRIDES
            : DEPROVISIONING_COMPONENTS_OVERRIDES
        )[status]
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
