import { useMemo } from 'react'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { EmptyState } from '@/components/common/EmptyState'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import { getInputDisplayName } from '@/utils/install-utils'
import type { TStepChangeCounts, TStepChangeSummary, TWorkflow } from '@/types'

type TChangedInput = {
  name: string
  old: string
  new: string
}

const inputCounts = (oldValue: string, newValue: string): TStepChangeCounts => {
  const created = oldValue === '(empty)' && newValue !== '(empty)'
  const removed = oldValue !== '(empty)' && newValue === '(empty)'
  return {
    create: created ? 1 : 0,
    update: !created && !removed ? 1 : 0,
    delete: removed ? 1 : 0,
    replace: 0,
    noop: 0,
  }
}

export const inputChangeSummaries = (
  workflow?: TWorkflow
): TStepChangeSummary[] => {
  const changes = parseWorkflowInputChanges(workflow)
  if (!changes.length) return []

  const counts = changes.reduce<TStepChangeCounts>(
    (total, change) => {
      const next = inputCounts(change.old, change.new)
      return {
        create: total.create + next.create,
        update: total.update + next.update,
        delete: total.delete + next.delete,
        replace: 0,
        noop: 0,
      }
    },
    { create: 0, update: 0, delete: 0, replace: 0, noop: 0 }
  )

  return [
    {
      stepId: 'install-inputs',
      stepName: 'Install inputs',
      planType: 'install_input',
      status: 'applied',
      counts,
      countsState: 'ok',
      hasDetail: true,
    },
  ]
}

export const parseWorkflowInputChanges = (
  workflow?: TWorkflow
): TChangedInput[] => {
  if (!workflow?.metadata?.changed_input_values) {
    return []
  }

  try {
    const parsed = JSON.parse(workflow.metadata.changed_input_values) as Record<
      string,
      { old?: string; new?: string }
    >
    return Object.entries(parsed).map(([name, values]) => ({
      name: getInputDisplayName(name),
      old: values.old || '(empty)',
      new: values.new || '(empty)',
    }))
  } catch {
    return []
  }
}

export const WorkflowInputChanges = ({
  divider = false,
  workflow,
  empty = false,
}: {
  divider?: boolean
  workflow: TWorkflow
  empty?: boolean
}) => {
  const changes = useMemo(() => parseWorkflowInputChanges(workflow), [workflow])

  if (!changes.length) {
    return empty ? (
      <EmptyState
        variant="table"
        emptyTitle="No input changes"
        emptyMessage="This input update has no recorded value changes."
      />
    ) : null
  }

  const content = (
    <PropertyGrid
      values={changes}
      columns={[
        { key: 'name', header: 'Input' },
        {
          key: 'old',
          header: 'Previous value',
          render: (value) => (
            <ClickToCopy>
              <Text
                className="line-through"
                variant="subtext"
                theme="error"
                family="mono"
              >
                {String(value)}
              </Text>
            </ClickToCopy>
          ),
        },
        {
          key: 'new',
          header: 'New value',
          render: (value) => (
            <ClickToCopy className="self-start">
              <Text variant="subtext" theme="success" family="mono">
                {String(value)}
              </Text>
            </ClickToCopy>
          ),
        },
      ]}
      gridTemplate="auto 1fr 1fr"
    />
  )

  return divider ? (
    <>
      <hr className="-mx-4" />
      {content}
    </>
  ) : (
    content
  )
}
