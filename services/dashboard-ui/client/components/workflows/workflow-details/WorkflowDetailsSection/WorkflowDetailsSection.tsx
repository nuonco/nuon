import { useState } from 'react'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { Modal } from '@/components/surfaces/Modal'
import { humanize } from '@/utils/string-utils'
import { getAccountDisplayName } from '@/utils/workflow-utils'
import type { TWorkflow, TInstall } from '@/types'
import { WorkflowInputChanges } from '../../WorkflowInputChanges'
import { WorkflowMetadata } from '../WorkflowMetadata'

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
  const [expanded, setExpanded] = useState(true)
  const toggleExpanded = () => setExpanded((prev) => !prev)
  const createdByName = getAccountDisplayName(workflow?.created_by)
  const createdByEmail = workflow?.created_by?.email

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
          {createdByEmail && createdByEmail !== createdByName ? (
            <Tooltip
              position="top"
              tipContent={
                <Text variant="subtext" family="mono">
                  {createdByEmail}
                </Text>
              }
              onClick={(e) => e.stopPropagation()}
            >
              <Text variant="base" weight="strong">
                {createdByName}
              </Text>
            </Tooltip>
          ) : (
            <Text variant="base" weight="strong">
              {createdByName}
            </Text>
          )}
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
                <Link href={`/${orgId}/apps/${install.app_id}`}>
                  {install?.app?.name}
                </Link>
              </LabeledValue>
            )}

            <LabeledValue label="Trigger">
              {humanize(workflow.type)}
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

      {expanded ? <WorkflowInputChanges divider workflow={workflow} /> : null}
    </Card>
  )
}
