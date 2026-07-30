import {
  importRunbook,
  operationLabels,
  serializeRunbook,
  validateStep,
  type TBuilderOperation,
  type TBuilderStep,
  type TOption,
} from '@/components/runbooks/RunbookBuilder/helpers'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks/get-runbooks'
import type { TCell, TStepCell } from './types'

export const cellSteps = (cells: TCell[]) =>
  cells.flatMap((cell) => (cell.kind === 'step' ? [cell.step] : []))

export const notebookHasReadme = (cells: TCell[]) =>
  cells.some((cell) => cell.kind === 'markdown' && cell.content.trim())

export const notebookToml = (
  name: string,
  description: string,
  cells: TCell[]
) =>
  serializeRunbook(
    name,
    description,
    cellSteps(cells),
    notebookHasReadme(cells)
  )

export function serializeNotebookReadme(
  name: string,
  description: string,
  cells: TCell[]
) {
  const heading = (value: string) => value.trim().replace(/\s+/g, ' ')
  const sections = [`# ${heading(name)}`]
  if (description.trim()) sections.push(description.trim())
  let stepIndex = 0
  cells.forEach((cell) => {
    if (cell.kind === 'markdown') {
      if (cell.content.trim()) sections.push(cell.content.trim())
      return
    }
    stepIndex += 1
    sections.push(
      `## ${stepIndex}. ${heading(cell.step.name)}`,
      `**Operation:** ${operationLabels[cell.step.operation]}`
    )
  })
  return `${sections.join('\n\n')}\n`
}

export function validateNotebook(name: string, cells: TCell[]) {
  const steps = cellSteps(cells)
  const errors: string[] = []
  const byCell: Record<string, string[]> = {}
  if (!name.trim()) errors.push('Enter a runbook name.')
  if (!steps.length) errors.push('Add at least one step.')
  let stepIndex = 0
  cells.forEach((cell) => {
    if (cell.kind !== 'step') return
    stepIndex += 1
    const stepErrors = validateStep(cell.step)
    if (stepErrors.length) {
      byCell[cell.key] = stepErrors.map((error) => `This step ${error}`)
      errors.push(...stepErrors.map((error) => `Step ${stepIndex} ${error}`))
    }
  })
  return { errors, byCell }
}

export const newMarkdownCell = (content = ''): TCell => ({
  key: crypto.randomUUID(),
  kind: 'markdown',
  content,
})

export const newStepCell = (operation: TBuilderOperation): TStepCell => ({
  key: crypto.randomUUID(),
  kind: 'step',
  step: {
    key: crypto.randomUUID(),
    operation,
    name: operationLabels[operation],
  },
})

export function stepTarget(step: TBuilderStep) {
  if (step.componentName) return step.componentName
  if (step.actionName) return step.actionName
  if (step.command) return step.command
  if (
    [
      'reprovision-sandbox',
      'check-sandbox-drift',
      'deprovision-sandbox',
    ].includes(step.operation)
  )
    return 'sandbox'
  return undefined
}

export function importRunbookCells(runbook: TRunbook, actions: TOption[]) {
  const result = importRunbook(runbook, actions)
  const cells = result.steps.flatMap((step): TCell[] => {
    const stepCell: TStepCell = {
      key: crypto.randomUUID(),
      kind: 'step',
      step: { ...step, documentation: undefined },
    }
    return step.documentation?.trim()
      ? [stepCell, newMarkdownCell(step.documentation.trim())]
      : [stepCell]
  })
  return { cells, errors: result.errors }
}

export const seedCells = (): TCell[] => [
  newMarkdownCell(
    'Walk through this runbook before running it against **{{.nuon.install.name}}**. Each step below runs in order.'
  ),
  newStepCell('check-component-drift'),
  newMarkdownCell(
    '> Review the plan output above before continuing. Nothing has been applied yet.'
  ),
  newStepCell('deploy-component'),
  newMarkdownCell(
    '## Verify\n\nConfirm https://{{.nuon.domain.public_domain}} is healthy before closing this out.'
  ),
]
