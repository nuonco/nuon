'use client'

import { useUser } from '@auth0/nextjs-auth0'
import { ArrowSquareOutIcon } from '@phosphor-icons/react'
import { Link } from '@/components/old/Link'
import { Text } from '@/components/old/Typography'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallWorkflow } from '@/types'

interface IInstallWorkflowActivity {
  installWorkflow: TInstallWorkflow
}

export const InstallWorkflowActivity = ({
  installWorkflow,
}: IInstallWorkflowActivity) => {
  const { user, isLoading } = useUser()
  const temporalLinkParams = useQueryParams({
    query: `\`WorkflowId\` STARTS_WITH "${installWorkflow?.owner_id}-execute-workflow-${installWorkflow?.id}"`,
  })
  const workflowSteps =
    installWorkflow?.steps?.filter((s) => s?.execution_type !== 'hidden') || []

  return (
    <div className="">
      <span className="flex w-full justify-between flex-wrap">
        <span className="flex flex-col gap-0">
          <span className="flex items-center gap-4 ml-auto">
            <span>&#x1F680;</span>
            <progress
              className="rounded-lg [&::-webkit-progress-bar]:rounded-lg [&::-webkit-progress-value]:rounded-lg   [&::-webkit-progress-bar]:bg-cool-grey-300 [&::-webkit-progress-value]:bg-green-400 [&::-moz-progress-bar]:bg-green-400 [&::-webkit-progress-value]:transition-all [&::-webkit-progress-value]:duration-500 [&::-moz-progress-bar]:transition-all [&::-moz-progress-bar]:duration-500 h-[8px]"
              max={workflowSteps.length}
              value={
                workflowSteps.filter(
                  (s) =>
                    s?.status?.status === 'success' ||
                    s?.status?.status === 'active' ||
                    s?.status?.status === 'error' ||
                    s?.status?.status === 'approved'
                ).length
              }
            />
          </span>

          <Text
            variant="reg-12"
            className="text-cool-grey-600 dark:text-white/70 self-end"
          >
            {
              workflowSteps.filter(
                (s) =>
                  s?.status?.status === 'success' ||
                  s?.status?.status === 'active' ||
                  s?.status?.status === 'error' ||
                  s?.status?.status === 'approved'
              ).length
            }{' '}
            of {workflowSteps.length} steps completed{' '}
            {workflowSteps.filter((s) => s?.status?.status === 'discarded')
              .length ? (
              <>
                ,{' '}
                {
                  workflowSteps.filter((s) => s?.status?.status === 'discarded')
                    .length
                }{' '}
                steps discarded
              </>
            ) : null}
          </Text>
        </span>
      </span>

      {!isLoading && user?.email?.endsWith('@nuon.co') ? (
        <Link
          className="text-base gap-2 mt-3 ml-auto"
          href={`/admin/temporal/namespaces/installs/workflows${temporalLinkParams}`}
          target="_blank"
        >
          View in Temporal <ArrowSquareOutIcon />
        </Link>
      ) : null}
    </div>
  )
}
