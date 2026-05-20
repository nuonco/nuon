import { Link, useParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { WorkflowActionButtons } from '../WorkflowActionButtons'
import type { TWorkflow, TInstall } from '@/types'

interface IWorkflowHeader {
  workflow: TWorkflow
  install?: TInstall
}

export const WorkflowHeader = ({ workflow, install }: IWorkflowHeader) => {
  const { orgId, installId } = useParams()
  const hasDrift =
    install?.drifted_objects?.length &&
    install?.drifted_objects?.find(
      (d) => d?.install_workflow_id === workflow?.id
    )

  return (
    <div className="flex flex-wrap items-center gap-3 justify-between w-full">
      <div className="flex flex-col gap-4">
        <Link
          to={`/${orgId}/installs/${installId}/workflows`}
          className="flex items-center gap-1.5 w-fit text-base leading-6 tracking-[-0.2px] font-strong text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 focus:text-primary-800 focus:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600 focus-visible:rounded"
        >
          <Icon variant="CaretLeftIcon" weight="bold" /> All workflows
        </Link>
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
            {workflow?.approval_option === 'approve-all' &&
            workflow?.metadata?.approval_type ? (
              <Badge variant="code" size="sm">
                {workflow.metadata.approval_type === 'approve-workflow'
                  ? 'auto-approve (workflow)'
                  : workflow.metadata.approval_type === 'install-config'
                    ? 'auto-approve (config)'
                    : 'auto-approve'}
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
