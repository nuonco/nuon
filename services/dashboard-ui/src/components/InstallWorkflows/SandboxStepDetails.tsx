'use client'

import { CaretRightIcon } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'

import type { TSandboxRun } from '@/types'
import type { IPollStepDetails } from './InstallWorkflowSteps'

export const SandboxStepDetails = ({
  step,
  shouldPoll = false,
  pollInterval = 5000,
}: IPollStepDetails) => {
  const { org } = useOrg()
  const {
    data: sandboxRun,
    isLoading,
    error,
  } = usePolling<TSandboxRun>({
    initIsLoading: true,
    path: `/api/orgs/${org.id}/installs/${step?.owner_id}/runs/${step?.step_target_id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <>
      {isLoading ? (
        <div className="border rounded-md p-6">
          <Loading loadingText="Loading sandobx details..." variant="stack" />
        </div>
      ) : (
        <>
          {error?.error ? (
            <Notice>
              {error?.error || 'Unable to load sandbox run details.'}
            </Notice>
          ) : null}
          {sandboxRun ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14">
                  Install sandbox {sandboxRun?.run_type}
                </Text>
                <Link
                  className="text-sm gap-0"
                  href={`/${org.id}/installs/${step?.owner_id}/sandbox/${sandboxRun?.id}`}
                >
                  View details
                  <CaretRightIcon />
                </Link>
              </div>
              <div className="p-6">
                <span className="flex gap-4 items-center">
                  <StatusBadge
                    description={
                      sandboxRun?.status_v2?.status_human_description ||
                      sandboxRun?.status_description
                    }
                    status={sandboxRun?.status_v2?.status || sandboxRun?.status}
                    label="Sandbox run status"
                  />
                </span>
              </div>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}
