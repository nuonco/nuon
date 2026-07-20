import type {
  TRunbook,
  TRunbookStep,
} from '@/lib/ctl-api/apps/runbooks/get-runbooks'

export type TBuilderOperation =
  | 'deploy-component'
  | 'check-component-drift'
  | 'tear-down-component'
  | 'reprovision-sandbox'
  | 'check-sandbox-drift'
  | 'deprovision-sandbox'
  | 'configured-action'
  | 'command'

export type TBuilderStep = {
  key: string
  operation: TBuilderOperation
  name: string
  documentation?: string
  componentName?: string
  deployDependents?: boolean
  tearDownDependents?: boolean
  skipComponentDeploys?: boolean
  actionName?: string
  command?: string
  timeout?: string
  role?: string
}

export type TOption = { id: string; name: string }

export const operationLabels: Record<TBuilderOperation, string> = {
  'deploy-component': 'Deploy component',
  'check-component-drift': 'Check component drift',
  'tear-down-component': 'Tear down component',
  'reprovision-sandbox': 'Reprovision sandbox',
  'check-sandbox-drift': 'Check sandbox drift',
  'deprovision-sandbox': 'Deprovision sandbox',
  'configured-action': 'Run configured action',
  command: 'Run command',
}

const escapeToml = (value: string) =>
  `"${value.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n').replace(/\r/g, '\\r').replace(/\t/g, '\\t')}"`

const goDurationPattern = /^[+-]?(?:0|(?:(?:\d+(?:\.\d+)?|\.\d+)(?:ns|us|µs|μs|ms|s|m|h))+)$/
const goDurationTermPattern = /(\d+(?:\.\d+)?|\.\d+)(ns|us|µs|μs|ms|s|m|h)/g
const durationUnitNanoseconds: Record<string, bigint> = {
  ns: BigInt(1),
  us: BigInt(1_000),
  µs: BigInt(1_000),
  μs: BigInt(1_000),
  ms: BigInt(1_000_000),
  s: BigInt(1_000_000_000),
  m: BigInt(60_000_000_000),
  h: BigInt(3_600_000_000_000),
}

function isGoDuration(value: string) {
  if (!goDurationPattern.test(value)) return false
  const unsigned = value.replace(/^[+-]/, '')
  if (unsigned === '0') return true
  let nanoseconds = BigInt(0)
  for (const match of unsigned.matchAll(goDurationTermPattern)) {
    const [whole = '0', fraction = ''] = match[1].split('.')
    const scale = BigInt(`1${'0'.repeat(fraction.length)}`)
    const numerator = BigInt(`${whole || '0'}${fraction}`)
    nanoseconds +=
      (numerator * durationUnitNanoseconds[match[2]]) / scale
  }
  const limit = value.startsWith('-')
    ? BigInt('9223372036854775808')
    : BigInt('9223372036854775807')
  return nanoseconds <= limit
}

export function validateRunbook(name: string, steps: TBuilderStep[]) {
  const errors: string[] = []
  if (!name.trim()) errors.push('Enter a runbook name.')
  if (!steps.length) errors.push('Add at least one step.')
  steps.forEach((step, index) => {
    const label = `Step ${index + 1}`
    if (!step.name.trim()) errors.push(`${label} requires a name.`)
    if (
      [
        'deploy-component',
        'check-component-drift',
        'tear-down-component',
      ].includes(step.operation) &&
      !step.componentName
    )
      errors.push(`${label} requires a component.`)
    if (step.operation === 'configured-action' && !step.actionName)
      errors.push(`${label} requires an action.`)
    if (step.operation === 'command' && !step.command?.trim())
      errors.push(`${label} requires a command.`)
    if (step.operation === 'command' && step.timeout?.trim() && !isGoDuration(step.timeout.trim()))
      errors.push(`${label} requires a valid timeout such as 30s or 5m.`)
  })
  return errors
}

export function serializeRunbook(
  name: string,
  description: string,
  steps: TBuilderStep[]
) {
  const lines = [`name = ${escapeToml(name.trim())}`]
  if (description.trim())
    lines.push(`description = ${escapeToml(description.trim())}`)
  if (steps.some((step) => step.documentation?.trim()))
    lines.push(
      `readme = ${escapeToml(`./runbooks/${slugifyRunbookName(name)}.md`)}`
    )
  steps.forEach((step) => {
    lines.push(
      '',
      '[[steps]]',
      `name = ${escapeToml(step.name.trim())}`
    )
    const componentDeploy = [
      'deploy-component',
      'check-component-drift',
    ].includes(step.operation)
    const type = componentDeploy
      ? 'component_deploy'
      : step.operation === 'tear-down-component'
        ? 'component_tear_down'
        : ['reprovision-sandbox', 'check-sandbox-drift'].includes(
              step.operation
            )
          ? 'sandbox_reprovision'
          : step.operation === 'deprovision-sandbox'
            ? 'sandbox_deprovision'
            : 'action'
    lines.push(`type = ${escapeToml(type)}`)
    if (
      ['check-component-drift', 'check-sandbox-drift'].includes(step.operation)
    )
      lines.push('plan_only = true')
    if (step.componentName)
      lines.push(`component_name = ${escapeToml(step.componentName)}`)
    if (step.deployDependents) lines.push('deploy_dependents = true')
    if (step.tearDownDependents) lines.push('tear_down_dependents = true')
    if (step.skipComponentDeploys) lines.push('skip_component_deploys = true')
    if (step.actionName)
      lines.push(`action_name = ${escapeToml(step.actionName)}`)
    if (step.command?.trim())
      lines.push(`command = ${escapeToml(step.command.trim())}`)
    if (step.timeout?.trim())
      lines.push(`timeout = ${escapeToml(step.timeout.trim())}`)
    if (step.role?.trim()) lines.push(`role = ${escapeToml(step.role.trim())}`)
  })
  return `${lines.join('\n')}\n`
}

export function serializeRunbookReadme(
  name: string,
  description: string,
  steps: TBuilderStep[]
) {
  const heading = (value: string) => value.trim().replace(/\s+/g, ' ')
  const sections = [`# ${heading(name)}`]
  if (description.trim()) sections.push(description.trim())
  steps.forEach((step, index) => {
    sections.push(
      `## ${index + 1}. ${heading(step.name)}`,
      `**Operation:** ${operationLabels[step.operation]}`
    )
    if (step.documentation?.trim()) sections.push(step.documentation.trim())
  })
  return `${sections.join('\n\n')}\n`
}

export const slugifyRunbookName = (name: string) =>
  name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '') || 'runbook'

function formatDuration(nanoseconds?: number) {
  if (!nanoseconds) return undefined
  const units = [
    ['h', 3_600_000_000_000],
    ['m', 60_000_000_000],
    ['s', 1_000_000_000],
    ['ms', 1_000_000],
  ] as const
  for (const [suffix, divisor] of units) {
    if (nanoseconds % divisor === 0) return `${nanoseconds / divisor}${suffix}`
  }
  return `${nanoseconds}ns`
}

export function importRunbook(runbook: TRunbook, actions: TOption[]) {
  const errors: string[] = []
  const actionById = new Map(actions.map((action) => [action.id, action.name]))
  const config = runbook.configs?.[0]
  if (config?.inputs?.length) {
    return {
      steps: [],
      errors: [`${runbook.name} defines runbook inputs and cannot be imported yet.`],
    }
  }
  const source =
    config?.steps
      ?.slice()
      .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []
  const steps = source.flatMap((step, index) => {
    const converted = convertImportedStep(
      step,
      actionById,
      `${runbook.name}, step ${index + 1}`,
      errors
    )
    return converted ? [converted] : []
  })
  return { steps: errors.length ? [] : steps, errors }
}

function convertImportedStep(
  step: TRunbookStep,
  actions: Map<string, string>,
  label: string,
  errors: string[]
): TBuilderStep | undefined {
  const hasInlineActionFields =
    !!step.command ||
    !!step.inline_contents ||
    !!step.role ||
    !!step.timeout ||
    !!(step.env_vars && Object.keys(step.env_vars).length)
  if (step.action_workflow_id && hasInlineActionFields) {
    errors.push(
      `${label} mixes a configured action with inline action fields and was skipped.`
    )
    return
  }
  let operation: TBuilderOperation | undefined
  if (step.type === 'component_deploy')
    operation = step.plan_only ? 'check-component-drift' : 'deploy-component'
  if (step.type === 'component_tear_down') operation = 'tear-down-component'
  if (step.type === 'sandbox_reprovision')
    operation = step.plan_only ? 'check-sandbox-drift' : 'reprovision-sandbox'
  if (step.type === 'sandbox_deprovision') operation = 'deprovision-sandbox'
  if (step.type === 'action' && step.command) operation = 'command'
  if (step.type === 'action' && step.inline_contents) {
    errors.push(`${label} uses inline script contents and was skipped.`)
    return
  }
  if (step.env_vars && Object.keys(step.env_vars).length) {
    errors.push(`${label} uses environment variables and was skipped.`)
    return
  }
  if (step.type === 'action' && !step.command) operation = 'configured-action'
  if (!operation) {
    errors.push(`${label} uses an unsupported operation and was skipped.`)
    return
  }
  const actionName = step.action_workflow_id
    ? actions.get(step.action_workflow_id)
    : undefined
  if (step.action_workflow_id && !actionName) {
    errors.push(`${label} could not map its configured action and was skipped.`)
    return
  }
  return {
    key: crypto.randomUUID(),
    operation,
    name: step.name || operationLabels[operation],
    componentName: step.component_name,
    deployDependents: step.deploy_dependents,
    tearDownDependents: step.tear_down_dependents,
    skipComponentDeploys: step.skip_component_deploys,
    actionName,
    command: step.command,
    timeout: formatDuration(step.timeout),
    role: step.role,
  }
}
