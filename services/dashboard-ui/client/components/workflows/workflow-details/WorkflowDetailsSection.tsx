'use client'

import { useMemo } from 'react'
import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { useInstall } from '@/hooks/use-install'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'

type ChangedInputValue = { old: string; new: string }

const ChangedInputsTooltipContent = ({
  changedInputValues,
}: {
  changedInputValues: string
}) => {
  const parsed = useMemo(() => {
    try {
      return JSON.parse(changedInputValues) as Record<
        string,
        ChangedInputValue
      >
    } catch {
      return null
    }
  }, [changedInputValues])

  if (!parsed || Object.keys(parsed).length === 0) return null

  return (
    <div className="flex flex-col gap-1 p-1 text-sm max-w-md">
      <Text variant="subtext" weight="strong">
        Changed inputs
      </Text>
      {Object.entries(parsed).map(([name, { old: oldVal, new: newVal }]) => (
        <div key={name} className="flex gap-1 font-mono text-xs">
          <span className="font-semibold">{name}:</span>
          <span className="opacity-60">{oldVal || '(empty)'}</span>
          <span>→</span>
          <span>{newVal || '(empty)'}</span>
        </div>
      ))}
    </div>
  )
}

export const WorkflowDetailsSection = () => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const { install } = useInstall()

  if (!workflow) return null

  const hasChangedInputs =
    workflow?.type === 'input_update' &&
    workflow?.metadata?.changed_input_values

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
          {hasChangedInputs ? (
            <Tooltip
              position="bottom"
              showIcon
              tipContent={
                <ChangedInputsTooltipContent
                  changedInputValues={
                    workflow.metadata!.changed_input_values!
                  }
                />
              }
              tipContentClassName="whitespace-normal"
            >
              {toSentenceCase(snakeToWords(workflow.type))}
            </Tooltip>
          ) : (
            toSentenceCase(snakeToWords(workflow.type))
          )}
        </LabeledValue>

        {install && (
          <LabeledValue label="App">
            <Text variant="subtext">
              <Link href={`/${org.id}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
          </LabeledValue>
        )}
      </div>
    </Expand>
  )
}