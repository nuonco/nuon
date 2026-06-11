import { useMemo, useState } from 'react'
import { Card } from '@/components/common/Card'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal } from '@/components/surfaces/Modal'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getInputDisplayName } from '@/utils/install-utils'
import type { TWorkflow, TInstall } from '@/types'
import { WorkflowMetadata } from '../WorkflowMetadata'

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

export const WorkflowDetailsSection = ({
  workflow,
  orgId,
  install,
}: IWorkflowDetailsSection) => {
  const changedInputs = useMemo<ChangedInput[]>(() => {
    if (
      workflow?.type !== 'input_update' ||
      !workflow?.metadata?.changed_input_values
    ) {
      return []
    }
    try {
      const parsed = JSON.parse(
        workflow.metadata.changed_input_values
      ) as Record<string, { old: string; new: string }>
      return Object.entries(parsed).map(
        ([name, { old: oldVal, new: newVal }]) => ({
          name: getInputDisplayName(name),
          old: oldVal || '(empty)',
          new: newVal || '(empty)',
        })
      )
    } catch {
      return []
    }
  }, [workflow?.type, workflow?.metadata?.changed_input_values])

  const [expanded, setExpanded] = useState(true)
  const toggleExpanded = () => setExpanded((prev) => !prev)

  return (
    <Card className="!p-4 !gap-4">
      <div
        role="button"
        tabIndex={0}
        aria-expanded={expanded}
        onClick={toggleExpanded}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            toggleExpanded()
          }
        }}
        className="flex items-center justify-between gap-3 cursor-pointer select-none focus:outline-none"
      >
        <div className="flex items-center gap-1.5 min-w-0">
          <Text variant="base" weight="strong">
            {workflow?.created_by?.email}
          </Text>
          <Text theme="neutral">
            initiated this workflow{' '}
            <Time time={workflow.created_at} format="relative" />
          </Text>
        </div>

        <span className="flex items-center gap-1 shrink-0 text-primary-600 dark:text-primary-400 text-xs font-strong tracking-tight">
          {expanded ? 'Show less' : 'Show more'}
          <Icon variant={expanded ? 'MinusIcon' : 'PlusIcon'} size="14" />
        </span>
      </div>

      {expanded && (
        <>
          <hr className="-mx-4" />

          <div className="flex flex-wrap items-start gap-x-16 gap-y-4">
            <LabeledValue label="Workflow ID">
              <ID theme="default">{workflow.id}</ID>
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

            <LabeledValue label="Trigger">
              {toSentenceCase(snakeToWords(workflow.type))}
            </LabeledValue>

            <div className="ml-auto">
              <Modal
                size="xl"
                heading="Metadata"
                className="h-[80vh]"
                triggerButton={{
                  variant: 'secondary',
                  size: 'sm',
                  className: 'shrink-0',
                  children: (
                    <>
                      <Icon variant="BracketsCurlyIcon" />
                      Metadata
                    </>
                  ),
                }}
              >
                <WorkflowMetadata workflow={workflow} />
              </Modal>
            </div>
          </div>
        </>
      )}

      {expanded && changedInputs.length > 0 && (
        <>
          <hr className="-mx-4" />

          <PropertyGrid
            values={changedInputs}
            columns={[
              { key: 'name', header: 'Name' },
              {
                key: 'old',
                header: 'Old value',
                render: (value) => (
                  <Text theme="neutral">
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
                  </Text>
                ),
              },
              {
                key: 'new',
                header: 'New value',
                render: (value) => (
                  <Text theme="neutral">
                    <ClickToCopy className="self-start">
                      <Text variant="subtext" theme="success" family="mono">
                        {String(value)}
                      </Text>
                    </ClickToCopy>
                  </Text>
                ),
              },
            ]}
            gridTemplate="auto 1fr 1fr"
          />
        </>
      )}
    </Card>
  )
}
