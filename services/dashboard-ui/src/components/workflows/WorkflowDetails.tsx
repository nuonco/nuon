'use client'

import { useUser } from '@auth0/nextjs-auth0'
import { ApproveAllButton } from '@/components/approvals/ApproveAll'
import { BackLink } from '@/components/common/BackLink'
import { Duration } from '@/components/common/Duration'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TApp, TInstall, TWorkflow } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { CancelWorkflowButton } from './CancelWorkflow'

import { Button } from '@/components/common/Button'

interface IWorkflowDetails extends IPollingProps {
  app?: TApp
  initWorkflow: TWorkflow
  install?: TInstall
}

export const WorkflowDetails = ({
  app,
  initWorkflow,
  install,
  pollInterval = 10000,
  shouldPoll = false,
}: IWorkflowDetails) => {
  const { user, isLoading } = useUser()
  const { org } = useOrg()
  const { data: workflow } = usePolling<TWorkflow>({
    initData: initWorkflow,
    path: `/api/orgs/${org.id}/workflows/${initWorkflow.id}`,
    pollInterval,
    shouldPoll,
  })
  const temporalLinkParams = useQueryParams({
    query: `\`WorkflowId\` STARTS_WITH "${workflow?.owner_id}-execute-workflow-${workflow?.id}"`,
  })
  const workflowSteps =
    workflow?.steps?.filter((s) => s?.execution_type !== 'hidden') || []

  return (
    <>
      <div className="flex flex-wrap items-center gap-3 justify-between w-full">
        <div className="flex flex-col gap-4">
          <BackLink />
          <HeadingGroup>
            <Text
              className="inline-flex gap-2 items-center"
              variant="h3"
              weight="strong"
            >
              {workflow.name || toSentenceCase(snakeToWords(workflow.type))}
            </Text>
            <Text theme="neutral">
              Watch your app get updated here and provide needed approvals.
            </Text>
          </HeadingGroup>
        </div>
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-4">
            {workflow?.approval_option === 'prompt' && !workflow?.finished && (
              <ApproveAllButton workflow={workflow} />
            )}
            {!workflow?.finished && (
              <CancelWorkflowButton workflow={workflow} />
            )}
            {!isLoading && user?.email?.endsWith('@nuon.co') ? (
              <Button
                href={`/admin/temporal/namespaces/installs/workflows${temporalLinkParams}`}
                target="_blank"
              >
                View in Temporal <Icon variant="ArrowSquareOutIcon" />
              </Button>
            ) : null}
          </div>
        </div>
      </div>

      <div className="flex flex-col md:flex-row gap-2 md:items-center justify-between">
        <div className="flex flex-wrap md:items-center gap-2 md:gap-6">
          <LabeledValue label="Elapsed time">
            <Duration nanoseconds={workflow?.execution_time} variant="base" />
          </LabeledValue>

          {workflow.plan_only ? (
            <LabeledValue label="Mode">
              <Tooltip
                position="right"
                showIcon
                tipContent={
                  <span className="flex flex-col w-66">
                    <Text weight="strong">Drift scan</Text>
                    <Text variant="subtext" className="text-nowrap">
                      Generate the workflow script without executing to detect
                      any drift between the app configuration and this install.
                    </Text>
                  </span>
                }
              >
                <Text variant="base">Drift scan</Text>
              </Tooltip>
            </LabeledValue>
          ) : null}
        </div>

        <div className="flex flex-wrap md:items-center gap-2 md:gap-6">
          {workflow.approval_option === 'prompt' && !workflow?.plan_only ? (
            <LabeledValue label="Pending approvals">
              <Text variant="base">
                {
                  workflowSteps.filter(
                    (s) =>
                      s?.execution_type === 'approval' &&
                      !s?.approval?.response &&
                      s?.status?.status !== 'discarded'
                  )?.length
                }
              </Text>
            </LabeledValue>
          ) : null}

          <LabeledValue label="Discarded">
            <Text variant="base">
              {
                workflowSteps.filter((s) => s?.status?.status === 'discarded')
                  .length
              }
            </Text>
          </LabeledValue>

          <LabeledValue label="Completed">
            <Text variant="base">
              {workflowSteps.filter((s) => s?.finished).length}
            </Text>
          </LabeledValue>

          <LabeledValue label="Total steps">
            <Text variant="base">{workflowSteps.length}</Text>
          </LabeledValue>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 md:gap-8 md:mt-6">
        <Text
          variant="h3"
          weight="stronger"
          className="inline-flex gap-2"
          theme={getStatusTheme(workflow.status.status) as any}
        >
          <Status status={workflow.status.status} variant="timeline" />
          {toSentenceCase(
            workflow.status.status_human_description || workflow.status.status
          )}
        </Text>

        <Text variant="h3" weight="stronger">
          Triggered via {snakeToWords(workflow.type)}
        </Text>
      </div>

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

          {install ? (
            <LabeledValue label="App">
              <Text variant="subtext">
                <Link href={`/${org.id}/apps/${install.app_id}`}>
                  {install?.app?.name}
                </Link>
              </Text>
            </LabeledValue>
          ) : null}
        </div>
      </Expand>
    </>
  )
}
