'use client'

import { useMemo } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useWorkflow } from '@/hooks/use-workflow'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { WorkflowActionButtons } from './WorkflowActionButtons'

type ChangedInputValue = { old: string; new: string }

const ChangedInputs = ({
  changedInputValues,
}: {
  changedInputValues: string
}) => {
  const parsed = useMemo(() => {
    try {
      return JSON.parse(changedInputValues) as Record<string, ChangedInputValue>
    } catch {
      return null
    }
  }, [changedInputValues])

  if (!parsed || Object.keys(parsed).length === 0) return null

  return (
    <div className="flex flex-col gap-1.5 mt-2">
      <Text variant="subtext" weight="strong" theme="neutral">
        Changed inputs
      </Text>
      <div className="flex flex-wrap gap-2">
        {Object.entries(parsed).map(([name, { old: oldVal, new: newVal }]) => (
          <Badge key={name} variant="code" size="sm">
            {name}: {oldVal || '(empty)'} → {newVal || '(empty)'}
          </Badge>
        ))}
      </div>
    </div>
  )
}

export const WorkflowHeader = () => {
  const { install } = useInstall()
  const { workflow } = useWorkflow()

  if (!workflow) return null

  return (
    <div className="flex flex-wrap items-center gap-3 justify-between w-full">
      <div className="flex flex-col gap-4">
        <BackLink />
        <HeadingGroup>
          <Text
            className="inline-flex gap-2 items-center"
            variant="h3"
            weight="strong"
          >
            {workflow?.type === 'action_workflow_run' &&
            workflow?.metadata?.adhoc_action
              ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
              : workflow.name || toSentenceCase(snakeToWords(workflow.type))}

            {install?.drifted_objects?.length &&
            install?.drifted_objects?.find(
              (d) => d?.install_workflow_id === workflow?.id
            ) ? (
              <Badge variant="code" theme="warn" size="sm">
                drift detected
              </Badge>
            ) : null}
          </Text>
          <Text theme="neutral">
            Watch your app get updated here and provide needed approvals.
          </Text>
          {workflow?.type === 'input_update' &&
            workflow?.metadata?.changed_input_values && (
              <ChangedInputs
                changedInputValues={workflow.metadata.changed_input_values}
              />
            )}
        </HeadingGroup>
      </div>

      <div className="flex flex-col gap-4">
        <WorkflowActionButtons />
      </div>
    </div>
  )
}
