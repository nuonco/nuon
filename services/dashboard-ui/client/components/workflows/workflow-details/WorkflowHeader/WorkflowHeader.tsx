import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { WorkflowActionButtons } from '../WorkflowActionButtons'
import type { TWorkflow, TInstall } from '@/types'

interface IWorkflowHeader {
  workflow: TWorkflow
  install?: TInstall
}

export const WorkflowHeader = ({ workflow, install }: IWorkflowHeader) => {
  const hasDrift =
    install?.drifted_objects?.length &&
    install?.drifted_objects?.find(
      (d) => d?.install_workflow_id === workflow?.id
    )

  return (
    <div className="flex flex-wrap items-center gap-3 justify-between w-full">
      <div className="flex flex-col gap-4">
        <BackLink />
        <HeadingGroup>
          <Text
            flex
            className="gap-2"
            variant="h3"
            weight="strong"
          >
            {workflow?.type === 'action_workflow_run' &&
            workflow?.metadata?.adhoc_action
              ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
              : workflow.name || toSentenceCase(snakeToWords(workflow.type))}

            {hasDrift ? (
              <Badge variant="code" theme="warn" size="sm">
                drift detected
              </Badge>
            ) : null}
          </Text>
          <Text theme="neutral">
            Watch your app get updated here and provide needed approvals.
          </Text>
        </HeadingGroup>
      </div>

      <div className="flex flex-col gap-4">
        <WorkflowActionButtons />
      </div>
    </div>
  )
}
