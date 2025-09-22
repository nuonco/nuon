'use client'

import { CaretRightIcon } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import type { TInstallDeploy } from '@/types'
import type { IPollStepDetails } from './InstallWorkflowSteps'

export const DeployStepDetails = ({
  pollInterval = 5000,
  step,
  shouldPoll = false,
}: IPollStepDetails) => {
  const { org } = useOrg()
  const {
    data: deploy,
    isLoading,
    error,
  } = usePolling<TInstallDeploy>({
    initIsLoading: true,
    path: `/api/orgs/${org.id}/installs/${step?.owner_id}/deploys/${step?.step_target_id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <>
      {step?.execution_type === 'approval' &&
      step?.status?.status === 'auto-skipped' ? (
        <div className="flex flex-col gap-2">
          <Text variant="reg-14">Plan had no changes, skipping deployent.</Text>
          <br></br>
        </div>
      ) : null}
      {isLoading ? (
        <div className="border rounded-md p-6">
          <Loading loadingText="Loading deploy details..." variant="stack" />
        </div>
      ) : (
        <>
          {error?.error ? (
            <Notice>{error?.error || 'Unable to load deploy'}</Notice>
          ) : null}
          {deploy ? (
            <div className="flex flex-col gap-8">
              <>
                {deploy ? (
                  <div className="flex flex-col border rounded-md shadow">
                    <div className="flex items-center justify-between p-3 border-b">
                      <Text variant="med-14">{deploy?.component_name}</Text>
                      <div className="flex items-center gap-4">
                        {deploy?.status !== 'queued' ? (
                          <Link
                            className="text-sm gap-0"
                            href={`/${org.id}/installs/${step?.owner_id}/components/${deploy?.component_id}`}
                          >
                            View component
                            <CaretRightIcon />
                          </Link>
                        ) : null}
                        {deploy?.status !== 'queued' ? (
                          <Link
                            className="text-sm gap-0"
                            href={`/${org.id}/installs/${step?.owner_id}/components/${deploy?.component_id}/deploys/${deploy?.id}`}
                          >
                            View deployment
                            <CaretRightIcon />
                          </Link>
                        ) : null}
                      </div>
                    </div>
                    <div className="p-6 flex flex-col gap-8">
                      <span className="flex gap-4 items-center">
                        <StatusBadge
                          description={
                            deploy?.status_v2?.status_human_description ||
                            deploy?.status
                          }
                          status={
                            deploy?.status_v2?.status ||
                            deploy?.status_description
                          }
                          label="Deployment status"
                        />
                      </span>
                    </div>
                  </div>
                ) : null}
              </>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}
