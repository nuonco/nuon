import { useMemo } from 'react'
import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import type { TWorkflow, TInstall } from '@/types'

type ChangedInput = {
  name: string
  old: string
  new: string
}

interface IWorkflowDetailsSection {
  workflow: TWorkflow
  orgId: string
  install?: TInstall
}

export const WorkflowDetailsSection = ({ workflow, orgId, install }: IWorkflowDetailsSection) => {
  const changedInputs = useMemo<ChangedInput[]>(() => {
    if (
      workflow?.type !== 'input_update' ||
      !workflow?.metadata?.changed_input_values
    ) {
      return []
    }
    try {
      const parsed = JSON.parse(workflow.metadata.changed_input_values) as Record<
        string,
        { old: string; new: string }
      >
      return Object.entries(parsed).map(([name, { old: oldVal, new: newVal }]) => ({
        name,
        old: oldVal || '(empty)',
        new: newVal || '(empty)',
      }))
    } catch {
      return []
    }
  }, [workflow?.type, workflow?.metadata?.changed_input_values])

  return (
    <Expand
      className="border rounded-md"
      id="workflow-details"
      isOpen
      heading={
        <span className="flex items-center gap-1.5">
          <Text variant="base" weight="strong">
            {workflow?.created_by?.email}
          </Text>
          <Text theme="neutral">
            initiated this workflow{' '}
            <Time time={workflow.created_at} format="relative" />
          </Text>
        </span>
      }
    >
      <div className="border-t flex flex-wrap items-center gap-6 md:gap-18 p-4">
        <LabeledValue label="Workflow ID">
          <ID theme="default">{workflow.id}</ID>
        </LabeledValue>

        <LabeledValue label="Trigger">
          {toSentenceCase(snakeToWords(workflow.type))}
        </LabeledValue>

        {install && (
          <LabeledValue label="App">
            <Text variant="subtext">
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
          </LabeledValue>
        )}

      </div>

      {changedInputs.length > 0 && (
        <div className="border-t p-4">
          <Text className="mb-2" variant="subtext" weight="strong" theme="neutral">
            Changed inputs
          </Text>
          <PropertyGrid
            values={changedInputs}
            columns={[
              { key: 'name', header: 'Name' },
              {
                key: 'old',
                header: 'Old value',
                render: (value) => (
                  <Text variant="subtext" theme="error" family="mono">
                    {String(value)}
                  </Text>
                ),
              },
              {
                key: 'new',
                header: 'New value',
                render: (value) => (
                  <Text variant="subtext" theme="success" family="mono">
                    {String(value)}
                  </Text>
                ),
              },
            ]}
            gridTemplate="auto auto 1fr"
          />
        </div>
      )}
    </Expand>
  )
}
