'use client'

import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useOrg } from '@/hooks/use-org'
import { useWorkflow } from '@/hooks/use-workflow'
import { useInstall } from '@/hooks/use-install'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'

export const WorkflowDetailsSection = () => {
  const { workflow } = useWorkflow()
  const { org } = useOrg()
  const { install } = useInstall()

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